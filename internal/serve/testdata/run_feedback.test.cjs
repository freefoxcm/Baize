const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const { join } = require('node:path');
const { test } = require('node:test');
const vm = require('node:vm');

const source = readFileSync(join(__dirname, '../assets/baize.js'), 'utf8');
function slice(start, end) {
  const from = source.indexOf(start);
  const to = source.indexOf(end, from);
  assert.ok(from >= 0 && to > from, `missing source boundaries: ${start}`);
  return source.slice(from, to);
}

function element(tag, className = '', text = '') {
  return {
    tag, className, textContent: text, children: [], attributes: {},
    classList: { add() {} },
    appendChild(child) { this.children.push(child); return child; },
    replaceChildren(...children) { this.children = children; },
    setAttribute(name, value) { this.attributes[name] = value; },
    get firstChild() { return this.children[0]; },
  };
}

function harness() {
  const region = element('div');
  const timers = new Map();
  let timerID = 0;
  const calls = [];
  const context = vm.createContext({
    $: () => region, el: element, __: key => key,
    setTimeout(fn, delay) { timers.set(++timerID, { fn, delay }); return timerID; },
    clearTimeout(id) { timers.delete(id); },
    EventSource: class {},
    historyPending: false, currentTurn: 0, turnArgChars: 0, todosDismissed: false,
    turnEls: new Map(), deliveryRecoveryActive: false,
    appendTranscriptNotice: text => calls.push(['notice', text]),
    showDeliveryReadiness: () => calls.push(['readiness']),
    log: { appendChild() { assert.fail('run error entered the transcript'); } },
  });
  for (const name of ['setConnState', 'clearRetrying', 'setRunning', 'clearPendingPrompts',
    'finalizeMsg', 'endModelActivity', 'autoSendGuidance', 'refreshWorkspaceAfterTurn',
    'loadSessions', 'fetchStatus', 'fetchTodos', 'refreshCheckpointAvailability', 'clearDeliveryCards']) {
    context[name] = () => calls.push([name]);
  }
  vm.runInContext(slice('let appToastTimer=0;', 'function hiddenTranscriptTool('), context);
  vm.runInContext('let es;\n' + slice('function connectEvents(){', '__authReady.then(connectEvents)'), context);
  context.connectEvents();
  const send = event => vm.runInContext(`es.onmessage({data:${JSON.stringify(JSON.stringify(event))}})`, context);
  return { context, region, timers, calls, send };
}

test('provider errors use a persistent alert with collapsed, text-only diagnostics', () => {
  const h = harness();
  const error = 'opencode-go: request failed: Post "https://opencode.ai/zen/go/v1/chat/completions": context canceled <img onerror=alert(1)>';
  h.send({ kind: 'turn_done', err: error });
  const toast = h.region.firstChild;
  assert.equal(toast.className, 'app-toast app-toast--danger');
  assert.equal(toast.attributes.role, 'alert');
  assert.equal(toast.children[0].textContent, 'turn_failed');
  const details = toast.children[0].children[0];
  assert.equal(details.tag, 'details');
  assert.equal(details.open, undefined);
  assert.equal(details.children[0].tag, 'summary');
  assert.equal(details.children[1].tag, 'pre');
  assert.equal(details.children[1].textContent, error);
  assert.equal(details.children[1].innerHTML, undefined);
  assert.equal(h.timers.size, 0);
  h.context.clearAppToast('connection');
  assert.equal(h.region.firstChild, toast);
  toast.children[1].onclick();
  assert.equal(h.region.children.length, 0);
});

test('only authoritative cancellation is shown as a transient stop', () => {
  for (const err of [undefined, 'provider: context canceled']) {
    const h = harness();
    h.send({ kind: 'turn_done', cancelled: true, err });
    const toast = h.region.firstChild;
    assert.equal(toast.className, 'app-toast app-toast--info');
    assert.equal(toast.children[0].textContent, 'turn_stopped');
    assert.equal(toast.children[0].children.length, 0);
    assert.equal([...h.timers.values()][0].delay, 3000);
  }
  for (const cancelled of [false, undefined, 'true']) {
    const h = harness();
    h.send({ kind: 'turn_done', cancelled, err: 'context canceled' });
    assert.equal(h.region.firstChild.className, 'app-toast app-toast--danger');
  }
});

test('success stays quiet, and a new turn clears only its old feedback', () => {
  const h = harness();
  h.send({ kind: 'turn_done' });
  assert.equal(h.region.children.length, 0);
  h.send({ kind: 'turn_done', err: 'HTTP 500' });
  h.send({ kind: 'turn_started' });
  assert.equal(h.region.children.length, 0);
  h.context.showAppToast('connection error', 'danger', 0, 'connection');
  h.send({ kind: 'turn_started' });
  assert.equal(h.region.firstChild.children[0].textContent, 'connection error');
});

test('structured recovery still uses its actionable card or notice', () => {
  for (const [outcome, expected] of [['final_readiness', 'readiness'], ['recovery_paused', 'notice']]) {
    const h = harness();
    h.send({ kind: 'turn_done', outcome, err: 'recovery detail' });
    assert.equal(h.region.children.length, 0);
    assert.ok(h.calls.some(([name]) => name === expected));
  }
});

test('late toast dismissal cannot remove a replacement toast', () => {
  const h = harness();
  h.context.showAppToast('first');
  [...h.timers.values()][0].fn();
  const dismissal = [...h.timers.values()].find(t => t.delay === 160).fn;
  h.send({ kind: 'turn_done', err: 'second' });
  dismissal();
  assert.equal(h.region.firstChild.children[0].textContent, 'turn_failed');
});
