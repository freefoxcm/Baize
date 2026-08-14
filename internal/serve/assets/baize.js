// ── vendor libraries (marked + DOMPurify + highlight.js) ──
const { marked, DOMPurify, hljs } = window.Vendor;
// ── i18n ──
const __LANG_PREF = document.documentElement.dataset.language || 'auto';
const __LANG = (__LANG_PREF === 'zh' || __LANG_PREF === 'en') ? __LANG_PREF : ((navigator.language || '').startsWith('zh') ? 'zh' : 'en');
document.documentElement.lang = __LANG;
const __T = {
  en: {
    'new_session': 'New Session',
    'compact': 'Compact',
    'rewind': 'Rewind',
    'branches': 'Branches',
    'models': 'Models',
    'sessions': 'Sessions',
    'status': 'Status',
    'loading': 'Loading...',
    'cache': 'Cache',
    'cost': 'Cost',
    'multi_currency': 'Multiple currencies',
    'balance': 'Balance',
    'ready': 'Ready',
    'thinking': 'Thinking...',
    'thinking_done': 'Thinking done',
    'retrying_status': 'Retrying ({attempt}/{max})...',
    'run_waiting_approval': 'Waiting for approval of {tool}',
    'run_waiting_ask': 'Waiting for your answer',
    'ask_title': 'Question',
    'ask_progress': 'Question {n}/{m}',
    'ask_stop_task': 'Stop task',
    'ask_next': 'Next',
    'ask_back': 'Back',
    'ask_custom_answer': 'Other answer',
    'ask_custom_answer_desc': 'Type your own answer',
    'ask_custom_placeholder': 'Type your own answer…',
    'ask_just_chat': 'Skip and keep chatting',
    'ask_just_chat_desc': 'Submit empty answers and continue the conversation.',
    'ask_select_hint': 'Use ↑↓ to navigate · Enter to confirm',
    'run_waiting_plan': 'Waiting for plan approval…',
    'approval_plan_badge': 'Plan',
    'approval_plan_title': 'Approve plan execution',
    'approval_plan_approve': 'Approve plan',
    'approval_plan_approve_desc': 'Run the approved plan',
    'run_announce_running': 'Thinking…',
    'modebar_title': 'Tool approval mode',
    'mode_ask': 'Ask',
    'mode_auto': 'Auto',
    'mode_yolo': 'YOLO',
    'mode_ask_title': 'Ask before each tool',
    'mode_auto_title': 'Auto-approve safe tools',
    'mode_yolo_title': 'Approve everything (YOLO)',
    'task_mode_title': 'Execution mode',
    'task_mode_direct': 'Standard',
    'task_mode_plan': 'Plan',
    'task_mode_goal': 'Goal',
    'work_mode_title': 'Work mode',
    'work_mode_balanced': 'Balanced',
    'work_mode_lightweight': 'Lightweight',
    'work_mode_delivery': 'Delivery',
    'model_search': 'Search models',
    'provider_label_deepseek': 'DeepSeek Official',
    'model_no_match': 'No matching models',
    'effort_title': 'Reasoning effort',
    'recovery_paused': 'Automatic retries paused. Baize stopped repeated attempts and kept completed work. Send “Continue” to start a fresh attempt, or add instructions to change direction.',
    'delivery_incomplete_title': 'Delivery checks are not complete',
    'delivery_incomplete_body': 'The response was generated, but verification and review still need to be completed.',
    'delivery_continue': 'Continue checks',
    'delivery_continue_prompt': 'Continue and complete the remaining delivery checks.',
    'delivery_missing': 'Still needed: {items}',
    'delivery_requirement_project_check': 'project checks',
    'delivery_requirement_todo': 'unfinished task items',
    'delivery_requirement_criteria': 'acceptance criteria',
    'delivery_requirement_verification': 'verification',
    'delivery_requirement_review': 'change review',
    'delivery_requirement_signoff': 'step sign-off',
    'delivery_requirement_action': 'host-observable work',
    'delivery_requirement_mutation': 'the requested change',
    'delivery_requirement_capability': 'required capability checks',
    'delivery_raw_detail': 'Technical detail',
    'todo_signable': 'sign off now',
    'todo_phase_signoff': 'phase sign-off',
    'stats': 'Stats',
    'statistics': 'Statistics',
    'model': 'Model',
    'workspace': 'Workspace',
    'workspace_files': 'Files',
    'workspace_search': 'Search workspace files',
    'workspace_refresh': 'Refresh files',
    'workspace_close': 'Close file panel',
    'workspace_resize': 'Resize file panel',
    'workspace_copy_path': 'Copy relative path',
    'workspace_select_file': 'Select a file to preview',
    'workspace_empty': 'No files',
    'workspace_no_results': 'No matching files',
    'workspace_load_error': 'Unable to load workspace files',
    'workspace_preview_error': 'Unable to preview this file',
    'workspace_decision_pending': 'Finish the current approval or question before opening files.',
    'workspace_binary': 'This binary file cannot be previewed.',
    'workspace_truncated': 'Preview limited to the first 2 MiB.',
    'workspace_source': 'Source',
    'workspace_report': 'Report',
    'workspace_pdf_previous': 'Previous page',
    'workspace_pdf_next': 'Next page',
    'workspace_pdf_zoom_out': 'Zoom out',
    'workspace_pdf_zoom_in': 'Zoom in',
    'workspace_pdf_fit_width': 'Fit width',
    'workspace_pdf_open': 'Open',
    'workspace_pdf_download': 'Download',
    'workspace_pdf_loading': 'Loading PDF…',
    'workspace_pdf_rendering': 'Rendering page…',
    'workspace_pdf_error': 'This PDF could not be previewed. You can still open or download it.',
    'workspace_pdf_protected': 'This PDF is password protected. Download it and open it in a trusted PDF reader.',
    'workspace_pdf_page': 'Page',
    'workspace_pdf_of': 'of',
    'sidebar_collapse': 'Collapse sidebar',
    'sidebar_expand': 'Expand sidebar',
    'total_tokens': 'Total Tokens',
    'cache_hit_rate': 'Cache Hit Rate',
    'total_cost': 'Session Cost',
    'context_usage': 'Context Usage',

    'connected': 'Connected',
    'reconnecting': 'Reconnecting...',
    'disconnected': 'Disconnected',
    'placeholder': 'Message Baize...  / for commands',
    'cmd_compact': 'Compact conversation',
    'cmd_new': 'New session',
    'cmd_clear': 'Clear context',
    'cmd_rewind': 'Rewind to checkpoint',
    'cmd_tree': 'Show branch tree',
    'cmd_branch': 'Create branch',
    'cmd_switch': 'Switch branch',
    'cmd_model': 'List/switch models',
    'cmd_provider': 'List/switch provider',
    'cmd_effort': 'Reasoning effort level',
    'cmd_mcp': 'MCP servers',
    'cmd_skill': 'Skills',
    'cmd_hooks': 'Hooks',
    'cmd_migrate': 'Migrate legacy data',
    'cmd_reload': 'Reload tools, skills, MCP and extensions',
    'cmd_reload_cmd': 'Reload commands',
    'cmd_plugins': 'Plugins',
    'cmd_memory': 'Show memory',
    'cmd_forget': 'Forget memory',
    'cmd_goal': 'Set a goal for the agent to pursue autonomously',
    'cmd_plan_exec': 'Execute plan todos',
    'cmd_prometheus': 'Planning interview',
    'cmd_thinking': 'Thinking effort',
    'cmd_help': 'Help',
    'cmd_docs': 'Local documentation',
    'cmd_group_session': 'Session',
    'cmd_group_branch': 'Branches',
    'cmd_group_model': 'Model',
    'cmd_group_system': 'System',
    'cmd_group_memory': 'Memory',
    'cmd_group_agent': 'Agent',
    'cmd_group_help': 'Help',
    'extensions_reloading': 'Reloading extensions...',
    'extensions_reloaded': 'Extensions reloaded',
    'extensions_reload_failed': 'Could not reload extensions',
    'command_palette': 'Command palette',
    'current': 'current',
    'command_empty': 'No commands match',
    'command_nav': '↑↓ navigate · Enter insert · Esc close',
    'danger': 'Danger',
    'tool_running': 'Running',
    'tool_done': 'Done',
    'work_working': 'Working…',
    'work_done': 'Worked',
    'tool_failed': 'Failed',
    'tool_copy': 'Copy output',
    'tool_copied': 'Copied',
    'tool_no_output': 'No output',
    'tool_args': 'Args',
    'tool_output': 'Output',
    'tool_audit': 'Execution audit',
    'tool_lines': 'lines',
    'tool_subcalls': 'sub-agent calls',
    'tool_blocked': 'Blocked',
    'tool_not_executed': 'not executed',
    'submit': 'Submit',
    'approval_title': 'Approval Required',
    'allow': 'Allow',
    'allow_once': 'Allow once',
    'allow_once_desc': 'Allow this operation once',
    'session_desc': 'Allow matching operations until this session ends',
    'persist_desc': 'Save as a persistent matching rule for future sessions',
    'deny_desc': 'Block this operation',
    'session': 'Allow for session',
    'session_command': 'Command for session',
    'session_prefix': 'Command prefix for session',
    'persist_command': 'Always allow command (save)',
    'persist_prefix': 'Always allow command prefix (save)',
    'persist_edit': 'Always allow edits (save)',
    'persist_tool': 'Always allow (save)',
    'session_edit': 'Edits for session',
    'session_tool': 'Tool for session',
    'deny': 'Deny',
    'auto_mode': 'Auto mode (auto-approve fallback approvals)',
    'plan_mode': 'Plan first; permissions and sandbox still apply',
    'yolo_mode': 'YOLO mode (auto-approve tool approvals; ask and plan still wait)',
    'goal_mode_desc': 'Goal mode — type a task for the agent to pursue autonomously',
    'goal_mode': 'Goal mode',
    'goal_placeholder': 'Describe your goal...',
    'goal_active': 'Active goal',
    'goal_exit': 'Exit goal mode',
    'auto': 'Auto',
    'plan': 'Plan',
    'goal_btn': 'Goal',
    'yolo': 'YOLO',
    'send': 'Send (Enter)',
    'cancel': 'Cancel (Esc)',
    'guidance_queue': 'Queued guidance',
    'guidance_count': 'Queued guidance #{n}',
    'guidance_remaining': '#{n} more queued',
    'guidance_collapse': 'Collapse',
    'guidance_mode': 'Guide',
    'guidance_send': 'Send this guidance to the transcript',
    'guidance_dismiss': 'Dismiss queued guidance',
    'guidance_steer_title': 'Queue as guidance',
    'placeholder_running': 'Running — type guidance, Enter adds it to the queue',
    'guidance_rejected': 'The turn ended before the guidance could be applied. It will be sent as a follow-up message.',
    'guidance_queued': 'Added to the guidance queue',
    'cancel_plain': 'Cancel',
    'delete': 'Delete',
    'delete_session': 'Delete Session',
    'cannot_delete_active': 'Cannot delete the active session',
    'delete_failed': 'Could not delete the session. Check your connection and try again.',
    'auth_failed': 'Web authentication failed',
    'search_sessions': 'Search sessions',
    'active': 'Active',
    'use_model': 'Use',
    'no_sessions': 'No sessions',
    'new_session_draft': 'New session',
    'session_draft': 'Draft',
    'subagent_reasoning': 'Reasoning summary',
    'subagent_response': 'Response preview',
    'subagent_notice': 'Activity',
    'subagent_result': 'Final result',
    'subagent_tools': 'Nested tools',
    'subagent_truncated': 'Preview truncated to the most recent content.',
    'subagent_phase_queued': 'Queued',
    'subagent_phase_running': 'Running',
    'subagent_phase_reasoning': 'Reasoning',
    'subagent_phase_responding': 'Responding',
    'subagent_phase_tool': 'Using tools',
    'subagent_phase_retrying': 'Retrying',
    'subagent_phase_completed': 'Completed',
    'subagent_phase_failed': 'Failed',
    'subagent_phase_cancelled': 'Cancelled',
    'settings': 'Settings',
    'close': 'Close',
    'settings_global_hint': 'Runtime settings are saved globally. Appearance stays in this browser.',
    'settings_models_agents': 'Models & agents',
    'settings_default_model': 'Default model',
    'settings_default_model_hint': 'Used when a new session is created.',
    'settings_planner_model': 'Planner model',
    'settings_subagent_model': 'Subagent model',
    'settings_subagent_effort': 'Subagent effort',
    'settings_execution': 'Execution',
    'settings_default_approval': 'Default approval mode',
    'settings_new_sessions_only': 'Applies to new sessions only.',
    'settings_subagent_depth': 'Subagent depth',
    'settings_subagent_concurrency': 'Total concurrency',
    'settings_parallel_writers': 'Parallel writers',
    'settings_compact_ratio': 'Compaction threshold',
    'settings_reasoning_language': 'Reasoning language',
    'settings_follow_conversation': 'Follow conversation',
    'settings_appearance': 'Appearance',
    'settings_theme': 'Theme',
    'settings_theme_auto': 'Follow system',
    'settings_theme_dark': 'Dark',
    'settings_theme_light': 'Light',
    'settings_density': 'Density',
    'settings_density_comfortable': 'Comfortable',
    'settings_density_compact': 'Compact',
    'settings_reasoning_display': 'Reasoning default',
    'settings_reasoning_auto': 'Open while running',
    'settings_reasoning_open': 'Always open',
    'settings_reasoning_closed': 'Always collapsed',
    'settings_subagent_preview': 'Subagent preview',
    'settings_preview_full': 'Full',
    'settings_preview_compact': 'Compact',
    'settings_subagent_collapse': 'Collapse subagents when completed',
    'settings_save': 'Save settings',
    'settings_retry': 'Retry apply',
    'settings_saved': 'Settings saved',
    'settings_pending': 'Saved; waiting for the current task to finish.',
    'settings_applying': 'Applying runtime settings…',
    'settings_applied': 'Runtime settings applied.',
    'settings_overridden': 'Some values are overridden by the current workspace: {fields}',
    'settings_conflict': 'Settings changed elsewhere. The latest values were loaded.',
    'settings_yolo_confirm': 'YOLO auto-approves tool permissions for new sessions. Save this default?',
    'error_loading': 'Error loading',
    'no_checkpoints': 'No checkpoints available.',
    'no_conversation': 'Start a conversation before using this action.',
    'no_branches': 'No conversation branches yet.',
    'switch': 'Switch',
    'compacted': 'Compacted',
    'messages': 'messages',
    'compacting': 'Compacting...',
    'scope_both': 'Code + conversation',
    'scope_conversation': 'Conversation only',
    'scope_code': 'Code only',
    'scope_fork': 'Fork (new branch)',
    'scope_sumfrom': 'Compress model context after here (visible history is kept)',
    'scope_sumupto': 'Compress model context before here (visible history is kept)',
    'delete_confirm': 'Delete this session?',
    'usage_calendar': 'Token activity',
    'cal_range_label': 'Token activity range',
    'cal_range_year': 'This year',
    'cal_range_6m': 'Last 6 months',
    'cal_range_3m': 'Last 3 months',
    'cal_less': 'less',
    'cal_more': 'more',
    'cal_total': '{t} tokens · {n} active days',
    'cal_loading': 'Loading token activity…',
    'cal_error': 'Token activity is temporarily unavailable.',
    'cal_turn': 'WebUI completed execution',
    'cal_turns': 'WebUI completed executions',
    'cal_scope_note': 'WebUI completed executions only; CLI and session interaction counts are separate.',
    'cal_scale_note': 'Color scale: active-day token P95 in this range',
    'cal_req': 'request',
    'cal_reqs': 'requests',
    'hint_commands': '/ commands',
    'hint_mode': 'Plan',
    'hint_yolo': 'YOLO',
    'hint_rewind': 'Esc×2 rewind',
    'question_nav_label': 'Question navigation',
    'question_nav_jump': 'Jump to question {n}',
    'example_explain': 'Explain the project structure',
    'example_fix': 'Find and fix any bugs',
    'example_test': 'Write tests for the main module',
    'theme_switch_light': 'Switch to light theme',
    'theme_switch_dark': 'Switch to dark theme',
    'new_session_busy': 'A task is running; start a new session after it finishes',
    'new_session_busy_title': 'Task in progress',
    'total': 'Total',
    'in': 'In',
    'out': 'Out',
    'nav_jk': 'j/k or ↑↓ to navigate',
    'nav_enter_esc': 'Enter to select · Esc to close',
    'nav_keys': 'b/c/d/f/s/u quick keys',
    'nav_apply_esc': 'Enter to apply · Esc to go back',
    'rewind_title': 'Rewind — Select Turn',
    'action_title': 'Turn #{turn} — Select Action',
    'files': 'files',
    'edit_message': 'Edit message',
    'edit_save': 'Save',
    'edit_cancel': 'Cancel edit',
    'edit_hint': 'Enter to save · Esc to cancel',
    'edited_note': 'edited',
    'approval_details': 'Details',
    'approval_hide': 'Hide details',
    'approval_confirm': 'Confirm',
    'approval_hint': '1-4 select · Enter confirm · Esc deny',
  },
  zh: {
    'new_session': '新会话',
    'compact': '压缩',
    'rewind': '回退',
    'branches': '分支',
    'models': '模型',
    'sessions': '会话',
    'status': '状态',
    'loading': '加载中...',
    'cache': '缓存',
    'cost': '费用',
    'multi_currency': '多币种',
    'balance': '余额',
    'ready': '就绪',
    'thinking': '思考中...',
    'thinking_done': '思考完成',
    'retrying_status': '正在重试 ({attempt}/{max})...',
    'run_waiting_approval': '等待审批:{tool}',
    'run_waiting_ask': '等待你的回答',
    'ask_title': '提问',
    'ask_progress': '问题 {n}/{m}',
    'ask_stop_task': '停止任务',
    'ask_next': '下一步',
    'ask_back': '上一步',
    'ask_custom_answer': '其他答案',
    'ask_custom_answer_desc': '输入你自己的答案',
    'ask_custom_placeholder': '输入你自己的答案…',
    'ask_just_chat': '跳过并继续聊天',
    'ask_just_chat_desc': '提交空答案并继续对话。',
    'ask_select_hint': '使用 ↑↓ 选择 · Enter 确认',
    'run_waiting_plan': '等待批准计划…',
    'approval_plan_badge': '计划',
    'approval_plan_title': '批准计划执行',
    'approval_plan_approve': '批准计划',
    'approval_plan_approve_desc': '按计划执行',
    'run_announce_running': '思考中…',
    'modebar_title': '工具审批模式',
    'mode_ask': '询问',
    'mode_auto': '自动',
    'mode_yolo': 'YOLO',
    'mode_ask_title': '每个工具执行前询问',
    'mode_auto_title': '自动批准安全工具',
    'mode_yolo_title': '全部批准 (YOLO)',
    'task_mode_title': '执行方式',
    'task_mode_direct': '常规',
    'task_mode_plan': '计划',
    'task_mode_goal': '目标',
    'work_mode_title': '工作模式',
    'work_mode_balanced': '均衡',
    'work_mode_lightweight': '轻量',
    'work_mode_delivery': '交付',
    'model_search': '搜索模型',
    'provider_label_deepseek': 'DeepSeek 官方',
    'model_no_match': '没有匹配的模型',
    'effort_title': '思考长度',
    'recovery_paused': '已暂停自动重试。Baize 已停止重复尝试，并保留已完成的工作。发送“继续”即可开始新一轮，也可以补充要求来调整方向。',
    'delivery_incomplete_title': '交付检查尚未完成',
    'delivery_incomplete_body': '内容已经生成，但验证和复审步骤尚未完成。',
    'delivery_continue': '继续检查',
    'delivery_continue_prompt': '继续完成剩余的交付检查。',
    'delivery_missing': '仍需完成：{items}',
    'delivery_requirement_project_check': '项目检查',
    'delivery_requirement_todo': '未完成待办',
    'delivery_requirement_criteria': '验收标准',
    'delivery_requirement_verification': '验证',
    'delivery_requirement_review': '变更复审',
    'delivery_requirement_signoff': '步骤签收',
    'delivery_requirement_action': '可观察的实际工作',
    'delivery_requirement_mutation': '用户要求的变更',
    'delivery_requirement_capability': '必需能力检查',
    'delivery_raw_detail': '技术详情',
    'todo_signable': '当前可签收',
    'todo_phase_signoff': '阶段签收',
    'stats': '统计',
    'statistics': '统计',
    'model': '模型',
    'workspace': '工作区',
    'workspace_files': '文件',
    'workspace_search': '搜索工作区文件',
    'workspace_refresh': '刷新文件',
    'workspace_close': '关闭文件面板',
    'workspace_resize': '调整文件面板宽度',
    'workspace_copy_path': '复制相对路径',
    'workspace_select_file': '选择一个文件进行预览',
    'workspace_empty': '没有文件',
    'workspace_no_results': '没有匹配的文件',
    'workspace_load_error': '无法加载工作区文件',
    'workspace_preview_error': '无法预览此文件',
    'workspace_decision_pending': '请先完成当前审批或问答，再打开文件。',
    'workspace_binary': '此二进制文件暂不支持预览。',
    'workspace_truncated': '预览仅显示前 2 MiB。',
    'workspace_source': '源码',
    'workspace_report': '报告',
    'workspace_pdf_previous': '上一页',
    'workspace_pdf_next': '下一页',
    'workspace_pdf_zoom_out': '缩小',
    'workspace_pdf_zoom_in': '放大',
    'workspace_pdf_fit_width': '适应宽度',
    'workspace_pdf_open': '打开',
    'workspace_pdf_download': '下载',
    'workspace_pdf_loading': '正在加载 PDF…',
    'workspace_pdf_rendering': '正在渲染页面…',
    'workspace_pdf_error': '无法预览此 PDF，仍可打开或下载文件。',
    'workspace_pdf_protected': '此 PDF 受密码保护，请下载后使用可信的 PDF 阅读器打开。',
    'workspace_pdf_page': '第',
    'workspace_pdf_of': '页，共',
    'sidebar_collapse': '收起侧栏',
    'sidebar_expand': '展开侧栏',
    'total_tokens': '总 Token',
    'cache_hit_rate': '缓存命中率',
    'total_cost': '会话费用',
    'context_usage': '上下文用量',

    'connected': '已连接',
    'reconnecting': '重新连接...',
    'disconnected': '已断开',
    'placeholder': '给 Baize 发消息...  / 查看命令',
    'cmd_compact': '压缩对话',
    'cmd_new': '新建会话',
    'cmd_clear': '清空上下文',
    'cmd_rewind': '回退到检查点',
    'cmd_tree': '显示分支树',
    'cmd_branch': '创建分支',
    'cmd_switch': '切换分支',
    'cmd_model': '列出/切换模型',
    'cmd_provider': '列出/切换服务商',
    'cmd_effort': '推理努力级别',
    'cmd_mcp': 'MCP 服务器',
    'cmd_skill': '技能',
    'cmd_hooks': '钩子',
    'cmd_migrate': '迁移旧版数据',
    'cmd_reload': '重新加载工具、技能、MCP 和扩展',
    'cmd_reload_cmd': '重新加载命令',
    'cmd_plugins': '插件',
    'cmd_memory': '显示记忆',
    'cmd_forget': '忘记记忆',
    'cmd_goal': '设置目标让代理自主执行',
    'cmd_plan_exec': '执行计划任务',
    'cmd_prometheus': '需求访谈规划',
    'cmd_thinking': '思考努力',
    'cmd_help': '帮助',
    'cmd_docs': '本地文档',
    'cmd_group_session': '会话',
    'cmd_group_branch': '分支',
    'cmd_group_model': '模型',
    'cmd_group_system': '系统',
    'cmd_group_memory': '记忆',
    'cmd_group_agent': '代理',
    'cmd_group_help': '帮助',
    'extensions_reloading': '正在重新加载扩展…',
    'extensions_reloaded': '扩展已重新加载',
    'extensions_reload_failed': '无法重新加载扩展',
    'command_palette': '命令面板',
    'current': '当前',
    'command_empty': '没有匹配的命令',
    'command_nav': '↑↓ 选择 · Enter 插入 · Esc 关闭',
    'danger': '危险',
    'tool_running': '运行中',
    'tool_done': '完成',
    'work_working': '工作中…',
    'work_done': '已完成',
    'tool_failed': '失败',
    'tool_copy': '复制输出',
    'tool_copied': '已复制',
    'tool_no_output': '无输出',
    'tool_args': '参数',
    'tool_output': '输出',
    'tool_audit': '运行审计',
    'tool_lines': '行',
    'tool_subcalls': '个子代理调用',
    'tool_blocked': '已阻止',
    'tool_not_executed': '未执行',
    'submit': '提交',
    'approval_title': '需要批准',
    'allow': '允许',
    'allow_once': '允许一次',
    'allow_once_desc': '仅允许此操作一次',
    'session_desc': '允许匹配操作,直到本会话结束',
    'persist_desc': '保存为持久匹配规则,后续会话记住',
    'deny_desc': '阻止此操作',
    'session': '本会话允许',
    'session_command': '本会话允许此命令',
    'session_prefix': '本会话允许命令前缀',
    'persist_command': '总是允许此命令（保存）',
    'persist_prefix': '总是允许命令前缀（保存）',
    'persist_edit': '总是允许编辑（保存）',
    'persist_tool': '总是允许（保存）',
    'session_edit': '本会话允许编辑',
    'session_tool': '本会话允许此工具',
    'deny': '拒绝',
    'auto_mode': '自动模式（自动批准兜底审批）',
    'plan_mode': '先规划；权限与沙箱仍然生效',
    'yolo_mode': 'YOLO 模式（自动批准工具权限；ask 和计划仍会等待）',
    'goal_mode_desc': '目标模式 — 输入任务让代理自主执行',
    'goal_mode': '目标模式',
    'goal_placeholder': '描述你的目标...',
    'goal_active': '活跃目标',
    'goal_exit': '退出目标模式',
    'auto': '自动',
    'plan': '计划',
    'goal_btn': '目标',
    'yolo': 'YOLO',
    'send': '发送 (Enter)',
    'cancel': '取消 (Esc)',
    'guidance_queue': '待处理引导',
    'guidance_count': '待处理引导 #{n}',
    'guidance_remaining': '还有 #{n} 条',
    'guidance_collapse': '收起',
    'guidance_mode': '引导',
    'guidance_send': '将这条引导加入信息流',
    'guidance_dismiss': '移除这条引导提示',
    'guidance_steer_title': '加入引导队列',
    'placeholder_running': '正在运行——输入补充指示，Enter 加入队列',
    'guidance_rejected': '本条引导未能应用，回合已结束。将作为后续消息发送。',
    'guidance_queued': '已加入引导队列',
    'cancel_plain': '取消',
    'delete': '删除',
    'delete_session': '删除会话',
    'cannot_delete_active': '无法删除当前会话',
    'delete_failed': '无法删除会话，请检查连接后重试',
    'auth_failed': 'Web 认证失败',
    'search_sessions': '搜索会话',
    'active': '当前',
    'use_model': '使用',
    'no_sessions': '暂无会话',
    'new_session_draft': '新会话',
    'session_draft': '草稿',
    'subagent_reasoning': '思考摘要',
    'subagent_response': '回答预览',
    'subagent_notice': '运行动态',
    'subagent_result': '最终结果',
    'subagent_tools': '子代理工具',
    'subagent_truncated': '预览已截断，仅保留最近内容。',
    'subagent_phase_queued': '排队中',
    'subagent_phase_running': '运行中',
    'subagent_phase_reasoning': '思考中',
    'subagent_phase_responding': '回答中',
    'subagent_phase_tool': '调用工具',
    'subagent_phase_retrying': '重试中',
    'subagent_phase_completed': '已完成',
    'subagent_phase_failed': '已失败',
    'subagent_phase_cancelled': '已取消',
    'settings': '设置',
    'close': '关闭',
    'settings_global_hint': '运行设置保存为用户全局配置，外观仅保存在当前浏览器。',
    'settings_models_agents': '模型与代理',
    'settings_default_model': '新会话默认模型',
    'settings_default_model_hint': '新建会话时使用。',
    'settings_planner_model': '规划模型',
    'settings_subagent_model': 'Subagent 模型',
    'settings_subagent_effort': 'Subagent 思考强度',
    'settings_execution': '执行',
    'settings_default_approval': '默认审批模式',
    'settings_new_sessions_only': '仅影响新会话。',
    'settings_subagent_depth': '子代理深度',
    'settings_subagent_concurrency': '总并发数',
    'settings_parallel_writers': '并行写入数',
    'settings_compact_ratio': '上下文压缩阈值',
    'settings_reasoning_language': '思考语言',
    'settings_follow_conversation': '跟随对话',
    'settings_appearance': '外观',
    'settings_theme': '主题',
    'settings_theme_auto': '跟随系统',
    'settings_theme_dark': '深色',
    'settings_theme_light': '浅色',
    'settings_density': '显示密度',
    'settings_density_comfortable': '舒适',
    'settings_density_compact': '紧凑',
    'settings_reasoning_display': '思考默认显示',
    'settings_reasoning_auto': '运行时展开',
    'settings_reasoning_open': '始终展开',
    'settings_reasoning_closed': '始终收起',
    'settings_subagent_preview': 'Subagent 预览',
    'settings_preview_full': '完整',
    'settings_preview_compact': '精简',
    'settings_subagent_collapse': '子代理完成后自动收起',
    'settings_save': '保存设置',
    'settings_retry': '重试应用',
    'settings_saved': '设置已保存',
    'settings_pending': '已保存，等待当前任务结束后应用。',
    'settings_applying': '正在应用运行设置…',
    'settings_applied': '运行设置已应用。',
    'settings_overridden': '当前工作区覆盖了部分设置：{fields}',
    'settings_conflict': '设置已被其他进程修改，已加载最新值。',
    'settings_yolo_confirm': 'YOLO 将对新会话自动批准工具权限。确认保存为默认模式吗？',
    'error_loading': '加载失败',
    'no_checkpoints': '暂无可用检查点',
    'no_conversation': '先开始一段对话，再使用此操作。',
    'no_branches': '暂无会话分支',
    'switch': '切换',
    'compacted': '已压缩',
    'messages': '条消息',
    'compacting': '压缩中...',
    'scope_both': '代码 + 对话',
    'scope_conversation': '仅对话',
    'scope_code': '仅代码',
    'scope_fork': '分叉（新分支）',
    'scope_sumfrom': '压缩此处之后的模型上下文（保留可见历史）',
    'scope_sumupto': '压缩此处之前的模型上下文（保留可见历史）',
    'delete_confirm': '删除此会话？',
    'usage_calendar': 'Token活动',
    'cal_range_label': 'Token活动范围',
    'cal_range_year': '全年',
    'cal_range_6m': '近6个月',
    'cal_range_3m': '近3个月',
    'cal_less': '少',
    'cal_more': '多',
    'cal_total': '{t} tokens · {n} 个活跃日',
    'cal_loading': '正在加载 Token活动…',
    'cal_error': 'Token活动暂时无法加载。',
    'cal_turn': '次 WebUI 完成执行',
    'cal_turns': '次 WebUI 完成执行',
    'cal_scope_note': '仅统计 WebUI 完成执行；不含 CLI，也不同于会话交互次数。',
    'cal_scale_note': '色阶：当前范围活跃日 Token P95',
    'cal_req': '请求',
    'cal_reqs': '请求',
    'hint_commands': '/ 命令',
    'hint_mode': '计划',
    'hint_yolo': 'YOLO',
    'hint_rewind': 'Esc×2 回退',
    'question_nav_label': '问题导航',
    'question_nav_jump': '跳转到问题 {n}',
    'example_explain': '解释项目结构',
    'example_fix': '查找并修复错误',
    'example_test': '为主模块编写测试',
    'theme_switch_light': '切换到浅色',
    'theme_switch_dark': '切换到深色',
    'new_session_busy': '任务进行中，无法新建会话',
    'new_session_busy_title': '任务进行中',
    'total': '总计',
    'in': '输入',
    'out': '输出',
    'nav_jk': 'j/k 或 ↑↓ 导航',
    'nav_enter_esc': 'Enter 选择 · Esc 关闭',
    'nav_keys': 'b/c/d/f/s/u 快捷键',
    'nav_apply_esc': 'Enter 应用 · Esc 返回',
    'rewind_title': '回退 — 选择轮次',
    'action_title': '第 #{turn} 轮 — 选择操作',
    'files': '个文件',
    'edit_message': '编辑消息',
    'edit_save': '保存',
    'edit_cancel': '取消编辑',
    'edit_hint': 'Enter 保存 · Esc 取消',
    'edited_note': '已编辑',
    'approval_details': '详情',
    'approval_hide': '隐藏详情',
    'approval_confirm': '确认',
    'approval_hint': '1-4 选择 · Enter 确认 · Esc 拒绝',
  }
};
const __ = (key, ...args) => {
  let s = __T[__LANG]?.[key] ?? __T.en[key] ?? key;
  if (args.length) args.forEach((v, i) => { s = s.replace('#{' + ['turn','n'][i] + '}', v); });
  return s;
};
function translateTokens(text) {
  return String(text).replace(/__\('([^']+)'\)/g, (_, key) => __(key));
}
function applyStaticI18n() {
  const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT, {
    acceptNode(node) {
      const parent = node.parentElement;
      if (!parent || parent.closest('script,style')) return NodeFilter.FILTER_REJECT;
      return node.nodeValue.includes("__('") ? NodeFilter.FILTER_ACCEPT : NodeFilter.FILTER_REJECT;
    }
  });
  const nodes = [];
  for (let node = walker.nextNode(); node; node = walker.nextNode()) nodes.push(node);
  nodes.forEach(node => { node.nodeValue = translateTokens(node.nodeValue); });
  document.querySelectorAll('[title],[placeholder],[aria-label]').forEach(node => {
    if (node.hasAttribute('title')) node.setAttribute('title', translateTokens(node.getAttribute('title')));
    if (node.hasAttribute('placeholder')) node.setAttribute('placeholder', translateTokens(node.getAttribute('placeholder')));
    if (node.hasAttribute('aria-label')) node.setAttribute('aria-label', translateTokens(node.getAttribute('aria-label')));
  });
}
applyStaticI18n();

// ── populate welcome examples ──
document.addEventListener('DOMContentLoaded', () => {
  document.querySelectorAll('.welcome__ex').forEach((btn, i) => {
    const keys = ['example_explain', 'example_fix', 'example_test'];
    if (i < keys.length) btn.dataset.prompt = __(keys[i]);
  });
});



const $ = s => document.querySelector(s);
const $$ = s => document.querySelectorAll(s);
const log = $('#log'), input = $('#in'), btnSend = $('#btn-send'), btnStop = $('#btn-stop');
const runStrip = $('#run-strip'), runStripText = $('#run-strip-text'), runStripAnnounce = $('#run-strip-announce');
const approvalSlot = $('#approval-slot');
const modebar = $('#modebar');
const statusDotSidebar = $('#status-dot'), statusModel = $('#status-model');
const ctxFill = $('#ctx-fill'), ctxUsed = $('#ctx-used'), ctxWindow = $('#ctx-window');
const welcome = $('#welcome');
const welcomeModel = $('#welcome-model'), welcomeCwd = $('#welcome-cwd');
const slashAnchor = $('#slash-menu-anchor');

// Token links use a URL fragment so the secret never appears in the initial
// HTTP request, browser history, referrers, or access logs. Exchange it for an
// HttpOnly cookie before any API fetch or SSE connection starts.
const __nativeFetch = window.fetch.bind(window);
function bootstrapFragmentToken() {
  const fragment = new URLSearchParams(window.location.hash.slice(1));
  const token = fragment.get('token');
  if(!token)return Promise.resolve();
  fragment.delete('token');
  const cleanHash=fragment.toString();
  window.history.replaceState(null,'',window.location.pathname+window.location.search+(cleanHash?'#'+cleanHash:''));
  return __nativeFetch('/auth/token',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({token})}).then(response=>{
    if(!response.ok)throw new Error(__('auth_failed')+' (HTTP '+response.status+')');
  });
}
const __authReady = bootstrapFragmentToken();
window.fetch = (...args) => __authReady.then(() => __nativeFetch(...args));

// ── state ──
let running = false, planMode = false, bypassMode = false, toolApprovalMode = 'ask', yoloRestoreMode = 'ask';
let goalMode = false, goalActive = false, goalText = '';
// guidance queue (desktop composer-guidance parity): mid-turn input is queued
// here while a turn runs, then steered in immediately or sent after turn_done.
let guidanceQueue = [];
let guidanceNextId = 1;
let guidanceExpanded = false;
let guidanceSendingId = null;
const GUIDANCE_VISIBLE = 2;
let turnStartAt = 0, turnTokens = 0, turnArgChars = 0, tickTimer = null, retryStatus = null;
let turnOutputTokens = 0, modelActiveAt = 0, modelActiveMs = 0; // desktop parity: per-turn output total + model-active window for TPS
let turnOutputChars = 0, turnOutputCharsAtUsage = 0; // desktop parity: live text+reasoning chars for the in-flight TPS estimate
let waitingPrompt = null, pendingApprovalLabel = '', decisionInteractionLocked = false;
let waitAccumMs = 0, waitStartedAt = 0; // desktop parity: ticker counts NET work time, waits (approval/ask) are excluded
function waitPause() { if (waitStartedAt === 0) waitStartedAt = Date.now(); }
function waitResume() { if (waitStartedAt !== 0) { waitAccumMs += Date.now() - waitStartedAt; waitStartedAt = 0; } }
let modelsCache = [];
let escTimer = null;
let hasVisibleHistory = false, checkpointCount = 0;
let sessionsCache = [], sessionFilter = '', pendingDeleteSession = null;

// ── slash commands registry ──
const SLASH_CMDS = [
  {cmd:'new',sig:'/new',desc:__('cmd_new'),group:'session'},
  {cmd:'clear',sig:'/clear',desc:__('cmd_clear'),group:'session'},
  {cmd:'compact',sig:'/compact',desc:__('cmd_compact'),group:'session'},
  {cmd:'rewind',sig:'/rewind',desc:__('cmd_rewind'),group:'session'},
  {cmd:'tree',sig:'/tree',desc:__('cmd_tree'),group:'branch'},
  {cmd:'branch',sig:'/branch <name>',desc:__('cmd_branch'),group:'branch'},
  {cmd:'switch',sig:'/switch <id>',desc:__('cmd_switch'),group:'branch'},
  {cmd:'model',sig:'/model [provider/model]',desc:__('cmd_model'),group:'model'},
  {cmd:'provider',sig:'/provider [name]',desc:__('cmd_provider'),group:'model'},
  {cmd:'effort',sig:'/effort <level>',desc:__('cmd_effort'),group:'model'},
  {cmd:'thinking',sig:'/thinking <level>',desc:__('cmd_thinking'),group:'model'},
  {cmd:'goal',sig:'/goal <task>',desc:__('cmd_goal'),group:'agent'},
  {cmd:'plan-exec',sig:'/plan-exec [--strict]',desc:__('cmd_plan_exec'),group:'agent'},
  {cmd:'prometheus',sig:'/prometheus <task>',desc:__('cmd_prometheus'),group:'agent'},
  {cmd:'mcp',sig:'/mcp',desc:__('cmd_mcp'),group:'system'},
  {cmd:'skill',sig:'/skill',desc:__('cmd_skill'),group:'system'},
  {cmd:'hooks',sig:'/hooks',desc:__('cmd_hooks'),group:'system'},
  {cmd:'migrate',sig:'/migrate [--from <dir>]',desc:__('cmd_migrate'),group:'system'},
  {cmd:'reload',sig:'/reload',desc:__('cmd_reload'),group:'system'},
  {cmd:'reload-cmd',sig:'/reload-cmd',desc:__('cmd_reload_cmd'),group:'system'},
  {cmd:'plugins',sig:'/plugins [list|show <name>]',desc:__('cmd_plugins'),group:'system'},
  {cmd:'memory',sig:'/memory',desc:__('cmd_memory'),group:'memory'},
  {cmd:'forget',sig:'/forget <item>',desc:__('cmd_forget'),group:'memory',danger:true},
  {cmd:'help',sig:'/help',desc:__('cmd_help'),group:'help'},
  {cmd:'docs',sig:'/docs [query]',desc:__('cmd_docs'),group:'help'},
];
const SLASH_GROUPS = ['session','branch','model','agent','system','memory','help'];
function slashGroupLabel(group){return __('cmd_group_'+group)||group;}

// ── helpers ──
function post(path, body) {
  return fetch(path, {method:'POST',headers:{'content-type':'application/json'},body:body?JSON.stringify(body):undefined});
}
// Follow streaming output only while the user is pinned to the bottom of the
// transcript; scrolling up detaches so history stays readable mid-stream.
// Detach triggers on wheel intent (any upward notch on scrollable content),
// not only on position: during fast streaming every frame re-bottoms the
// viewport, so a slow trackpad drag would never accumulate the 40px position
// delta. force re-pins (own message sent, session loaded, action prompts,
// click feedback) and wins over a concurrent user scroll — an approval needs
// a decision and its key handler is document-level, so it must not sit
// off-screen. Scrolls are instant: a smooth animation lags behind rapid
// chunks, and its intermediate positions would read as "not at bottom" and
// break pin detection.
let pinnedToBottom=true;
function atBottom() { return log.scrollHeight-log.scrollTop-log.clientHeight<40; }
log.addEventListener('scroll',()=>{pinnedToBottom=atBottom();});
log.addEventListener('wheel',e=>{if(e.deltaY<0&&log.scrollHeight>log.clientHeight)pinnedToBottom=false;},{passive:true});
function scrollDown(force) { if(force)pinnedToBottom=true; if(!pinnedToBottom)return; requestAnimationFrame(()=>{if(force)pinnedToBottom=true; else if(!pinnedToBottom)return; log.scrollTo({top:log.scrollHeight,behavior:'instant'});});}
function hideWelcome() { if(welcome) welcome.style.display='none';}
function showWelcome(){if(welcome)welcome.style.display='';setUsageCalendarRange('6m',true);}

// ── welcome Token activity (Codex-style week-column heatmap) ──
let calData=null, calRange='6m', calSelected=null, calAbort=null, calGeneration=0, calWeeks=0;
const CAL_RANGES=new Set(['year','6m','3m']);
function fmtDayKey(dt){return dt.getFullYear()+'-'+String(dt.getMonth()+1).padStart(2,'0')+'-'+String(dt.getDate()).padStart(2,'0');}
function parseDayKey(key){const p=String(key||'').split('-').map(Number);return new Date(p[0],p[1]-1,p[2]);}
function calLevel(toks,max){if(toks<=0)return 0;if(max<=0)return 1;return Math.max(1,Math.min(5,Math.ceil(toks/max*5)));}
function updateCalRangeButtons(){$$('[data-cal-range]').forEach(btn=>btn.setAttribute('aria-pressed',btn.dataset.calRange===calRange?'true':'false'));}
function setUsageCalendarRange(range,force){
  if(!CAL_RANGES.has(range))return;
  const changed=calRange!==range;calRange=range;updateCalRangeButtons();closeCalTip();
  if(changed||force||!calData)loadUsageCalendar();
}
function sizeUsageCalendarGrid(){
  const body=$('#cal-body'),grid=$('#cal-grid');if(!body||!grid||calWeeks<=0)return;
  const gap=3,available=Math.max(1,body.clientWidth-2);
  const size=Math.max(10,Math.min(14,Math.floor((available-Math.max(0,calWeeks-1)*gap)/calWeeks)));
  grid.style.setProperty('--cal-cell',size+'px');
}
function renderUsageCalendar(d){
  const grid=$('#cal-grid');if(!grid)return;grid.innerHTML='';
  const days=Array.isArray(d.days)?d.days:[];
  const first=days.length?parseDayKey(days[0].day):new Date();
  const offset=(first.getDay()+6)%7; // Monday = row 0
  calWeeks=Math.max(1,Math.ceil((days.length+offset)/7));
  grid.style.gridTemplateColumns='repeat('+calWeeks+',var(--cal-cell))';
  const todayKey=fmtDayKey(new Date());
  days.forEach((entry,i)=>{
    const dayKey=entry.day,toks=entry.tokens||0,turns=entry.turns||0;
    const row=(offset+i)%7+1,col=Math.floor((offset+i)/7)+1,lvl=Number.isInteger(entry.level)?Math.max(0,Math.min(5,entry.level)):calLevel(toks,d.max||0);
    const cell=el('button','cal-cell'+(lvl?' cal-cell--'+lvl:'')+(dayKey===todayKey?' cal-cell--today':''));
    cell.type='button';cell.dataset.day=dayKey;cell.style.gridRow=String(row);cell.style.gridColumn=String(col);
    cell.setAttribute('role','gridcell');cell.setAttribute('aria-expanded','false');
    cell.setAttribute('aria-label',dayKey+' · '+fmtTok(toks)+' tokens · '+turns+' '+(turns===1?__('cal_turn'):__('cal_turns')));
    cell.onmouseenter=()=>showCalTip(cell,dayKey,entry,false);cell.onmouseleave=()=>hideCalTip(false);
    cell.onfocus=()=>showCalTip(cell,dayKey,entry,false);cell.onblur=()=>hideCalTip(false);
    cell.onclick=e=>{e.stopPropagation();toggleCalDay(cell,dayKey,entry);};grid.appendChild(cell);
  });
  const sum=$('#cal-sum');if(sum)sum.textContent=__('cal_total').replace('{t}',fmtTok(d.total||0)).replace('{n}',String(d.activeDays||0));
  const scale=$('#cal-scale-note');if(scale){scale.textContent=__('cal_scale_note');scale.title=__('cal_scope_note');}
  sizeUsageCalendarGrid();
}
function loadUsageCalendar(){
  const cal=$('#welcome-calendar'),error=$('#cal-error'),sum=$('#cal-sum');if(!cal)return;
  if(calAbort)calAbort.abort();const generation=++calGeneration;calAbort=new AbortController();
  cal.classList.add('welcome__calendar--loading');cal.style.display='';
  if(error)error.style.display='none';if(sum)sum.textContent=__('cal_loading');
  fetch('/usage/calendar?range='+encodeURIComponent(calRange),{signal:calAbort.signal}).then(r=>{if(!r.ok)throw new Error('HTTP '+r.status);return r.json();}).then(d=>{
    if(generation!==calGeneration)return;calData=d;closeCalTip();renderUsageCalendar(d);
  }).catch(err=>{
    if(err&&err.name==='AbortError')return;if(generation!==calGeneration)return;calData=null;
    const grid=$('#cal-grid');if(grid)grid.innerHTML='';if(sum)sum.textContent='';
    if(error){error.textContent=__('cal_error');error.style.display='';}
  }).finally(()=>{if(generation===calGeneration)cal.classList.remove('welcome__calendar--loading');});
}
function calTipHTML(dayKey,entry){
  const e=entry||{},tokens=e.tokens||0,requests=e.requests||0,turns=e.turns||0;
  let html='<div><b>'+escHtml(dayKey)+'</b> · '+fmtTok(tokens)+' tokens · '+requests+' '+(requests===1?__('cal_req'):__('cal_reqs'))+' · '+turns+' '+(turns===1?__('cal_turn'):__('cal_turns'))+'</div>';
  html+='<div class="cal-method">'+escHtml(__('cal_scope_note'))+'</div>';
  const models=Object.entries(e.byModel||{}).sort((a,b)=>b[1]-a[1]);
  if(models.length)html+='<div class="cal-models">'+models.map(([m,t])=>'<div class="cal-model"><span>'+escHtml(m)+'</span><b>'+fmtTok(t)+'</b></div>').join('')+'</div>';
  return html;
}
function positionCalTip(cell){
  const tip=$('#cal-tip');if(!tip||tip.style.display==='none')return;
  const rect=cell.getBoundingClientRect(),half=Math.max(80,tip.offsetWidth/2+6);
  const left=Math.max(half,Math.min(window.innerWidth-half,rect.left+rect.width/2));
  const below=rect.top-tip.offsetHeight-8<8;tip.classList.toggle('welcome__calendar-tip--below',below);
  tip.style.left=left+'px';tip.style.top=(below?rect.bottom+8:rect.top-8)+'px';
}
function showCalTip(cell,dayKey,entry,sticky){
  if(calSelected&&!sticky)return;const tip=$('#cal-tip');if(!tip)return;
  tip.innerHTML=calTipHTML(dayKey,entry);tip.style.display='';tip.setAttribute('aria-hidden','false');tip.dataset.day=dayKey;
  if(sticky){calSelected=dayKey;document.querySelectorAll('.cal-cell--active').forEach(c=>{c.classList.remove('cal-cell--active');c.setAttribute('aria-expanded','false');});cell.classList.add('cal-cell--active');cell.setAttribute('aria-expanded','true');}
  positionCalTip(cell);
}
function hideCalTip(force){
  if(calSelected&&!force)return;const tip=$('#cal-tip');if(!tip)return;
  tip.style.display='none';tip.setAttribute('aria-hidden','true');tip.removeAttribute('data-day');
}
function closeCalTip(){calSelected=null;document.querySelectorAll('.cal-cell--active').forEach(c=>{c.classList.remove('cal-cell--active');c.setAttribute('aria-expanded','false');});hideCalTip(true);}
function toggleCalDay(cell,dayKey,entry){if(calSelected===dayKey){closeCalTip();return;}showCalTip(cell,dayKey,entry,true);}
$$('[data-cal-range]').forEach(btn=>btn.onclick=()=>setUsageCalendarRange(btn.dataset.calRange,false));
document.addEventListener('pointerdown',e=>{if(calSelected&&!e.target.closest('.cal-cell'))closeCalTip();});
document.addEventListener('keydown',e=>{if(e.key==='Escape'&&calSelected){e.preventDefault();e.stopImmediatePropagation();closeCalTip();}});
if(typeof ResizeObserver!=='undefined')new ResizeObserver(sizeUsageCalendarGrid).observe($('#cal-body'));
// Global single-key shortcuts must stay dead while the user is typing (the
// composer is autofocused, so this is the default state) and when a
// browser/OS chord is held (Cmd+A must never resolve an approval).
function isTextEntry(t) { return !!t&&(t.tagName==='INPUT'||t.tagName==='TEXTAREA'||t.tagName==='SELECT'||t.isContentEditable); }
function isPlainKey(e) { return !e.ctrlKey&&!e.metaKey&&!e.altKey&&!isTextEntry(e.target); }
// fmtTok mirrors desktop's formatTokens (desktop/frontend/src/lib/format.ts):
//  - n <= 0            → "-"
//  - n >= 1_000_000    → "X.YM" (trailing .0 stripped)
//  - n >= 1_000        → "X.YK" (trailing .0 stripped)
//  - otherwise         → plain number
function fmtTok(n) {
  if (typeof n !== 'number' || n <= 0) return '-';
  if (n >= 1000000) return (n / 1000000).toFixed(1).replace(/\.0$/, '') + 'M';
  if (n >= 1000) return (n / 1000).toFixed(1).replace(/\.0$/, '') + 'K';
  return String(n);
}
function workspaceDisplayName(path){
  const raw=String(path||'').trim();if(!raw)return '-';
  const trimmed=raw.replace(/[\\/]+$/,'');if(!trimmed)return raw;
  const parts=trimmed.split(/[\\/]/);return parts[parts.length-1]||raw;
}
function currencySymbol(c){const v=String(c||'¥').trim();if(/^(cny|rmb|yuan)$/i.test(v))return '¥';if(/^(usd|dollar)$/i.test(v))return '$';return v||'¥';}
function fmtMoney(n,c){if(typeof n!=='number'||!isFinite(n)||n<=0)return '—';const s=currencySymbol(c);return '≈'+s+(n<1?n.toFixed(4):n.toFixed(2));}
function usageSelectedCost(usage){
  if(!usage)return null;
  const quote=usage.costQuote;
  if(quote&&(quote.displayStatus==='bucketed'||quote.aggregateMode==='currency_buckets'))return{bucketed:true};
  const selected=quote?.selected;
  if(selected?.amount){const amount=Number(selected.amount);if(Number.isFinite(amount)&&amount>0)return{amount,currency:selected.currency||usage.currencyCode||usage.currency};}
  if(quote)return null;
  const amount=usage.cost??usage.costUsd;
  if(typeof amount==='number'&&amount>0)return{amount,currency:usage.currencyCode||usage.currency};
  return null;
}
function fmtElapsed(ms) {
  const s = Math.floor(Math.max(0, ms) / 1000); // desktop parity: whole seconds (12s / 1m 23s)
  if (s < 60) return s + 's';
  return Math.floor(s / 60) + 'm ' + (s % 60) + 's';
}
function escHtml(s) { return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');}
function escAttr(s) { return escHtml(s).replace(/"/g,'&quot;').replace(/'/g,'&#39;');}
function el(tag,cls,text) { const e=document.createElement(tag); if(cls)e.className=cls; if(text)e.textContent=text; return e;}
function compactText(s, max) {
  const text=String(s||'').replace(/\s+/g,' ').trim();
  return text.length>max?text.slice(0,max-1)+'…':text;
}
function lineCount(s) {
  const text=String(s||'');
  return text?text.split(/\r\n|\r|\n/).length:0;
}
function actionDisabled(id){return $('#'+id)?.classList.contains('sidebar__item--disabled');}
function setActionDisabled(id, disabled, title) {
  const node = $('#'+id);
  if(!node)return;
  node.classList.toggle('sidebar__item--disabled', !!disabled);
  node.setAttribute('aria-disabled', disabled ? 'true' : 'false');
  if(title)node.setAttribute('title', title);
  else node.removeAttribute('title');
}
function showNotice(text, tone) {
  hideWelcome();
  const n=el('div','notice'+(tone==='warn'?' notice--warn':''),text);
  log.appendChild(n);
  scrollDown(true); // synchronous feedback to a user action — always reveal
}
function hiddenTranscriptTool(name){
  const n=String(name||'').trim().toLowerCase();
  return n==='todo_write'||n==='exit_plan_mode';
}
const DELIVERY_REQUIREMENTS={
  project_check:'delivery_requirement_project_check',todo:'delivery_requirement_todo',criteria:'delivery_requirement_criteria',
  verification:'delivery_requirement_verification',review:'delivery_requirement_review',signoff:'delivery_requirement_signoff',
  action:'delivery_requirement_action',mutation:'delivery_requirement_mutation',capability:'delivery_requirement_capability'
};
function disableDeliveryCards(){
  $$('.delivery-card__button').forEach(btn=>{btn.disabled=true;});
}
function clearDeliveryCards(){
  $$('.delivery-card').forEach(card=>card.remove());
  items=items.filter(it=>it.kind!=='delivery');
}
function renderDeliveryCard(it){
  const card=el('div','delivery-card');
  card.dataset.deliveryId=it.id||'';
  card.appendChild(el('div','delivery-card__title',__('delivery_incomplete_title')));
  card.appendChild(el('div','delivery-card__body',__('delivery_incomplete_body')));
  if(it.detail)card.appendChild(el('div','delivery-card__detail',it.detail));
  if(it.raw){
    const details=el('details','delivery-card__raw');
    const summary=el('summary','',__('delivery_raw_detail'));
    details.appendChild(summary);details.appendChild(el('div','',it.raw));card.appendChild(details);
  }
  const actions=el('div','delivery-card__actions');
  const btn=el('button','delivery-card__button',__('delivery_continue'));
  btn.type='button';
  btn.onclick=()=>{
    if(btn.disabled||running)return;
    disableDeliveryCards();
    deliveryRecoveryActive=true;
    const prompt=__('delivery_continue_prompt');
    post('/delivery-recovery',{display:prompt,input:prompt}).then(async response=>{
      if(response.ok)return;
      throw new Error((await response.text()).trim()||('HTTP '+response.status));
    }).catch(err=>{
      deliveryRecoveryActive=false;
      btn.disabled=false;
      showNotice(String(err&&err.message||err), 'warn');
    });
  };
  actions.appendChild(btn);card.appendChild(actions);
  return card;
}
function showDeliveryReadiness(e){
  disableDeliveryCards();
  const missing=Array.isArray(e&&e.readiness&&e.readiness.missing)?e.readiness.missing:[];
  const labels=missing.map(id=>DELIVERY_REQUIREMENTS[id]).filter(Boolean).map(key=>__(key));
  const detail=labels.length?__('delivery_missing').replace('{items}',labels.join(__LANG==='zh'?'、':', ')):'';
  const raw=String(e&&e.err||'');
  const it={id:genItemId(),kind:'delivery',detail,raw,turn:currentTurn};
  items.push(it);
  const card=renderDeliveryCard(it);appendItem(it,card);scrollDown(true);
}
function updateActionAvailability() {
}

// ── run strip (desktop composer-run-strip) ──
const SPINNER_WORDS = {
  en: ['Frolicking','Pondering','Noodling','Brewing','Conjuring','Cogitating','Percolating','Ruminating','Simmering','Synthesizing','Tinkering','Marinating','Crunching','Hatching','Mulling','Whirring','Forging','Spelunking','Puttering','Vibing'],
  zh: ['梳理思路','聚焦问题','拆解任务','推演方案','权衡取舍','推敲细节','整合结论','评估进展','校验结果','检索资料','构建方案','验证逻辑','总结要点','打磨输出','迭代优化','排查障碍','规划步骤','分析上下文','整理成果','完善方案'],
};
function beginModelActivity() { if (!modelActiveAt) modelActiveAt = Date.now(); }
function endModelActivity() { if (modelActiveAt) { modelActiveMs += Date.now() - modelActiveAt; modelActiveAt = 0; } }
function tickerText() {
  const words = SPINNER_WORDS[__LANG] || SPINNER_WORDS.en;
  const ms = Math.max(0, Date.now() - turnStartAt - waitAccumMs - (waitStartedAt ? Date.now() - waitStartedAt : 0));
  const word = words[Math.floor(ms / 3000) % words.length];
  // Desktop parity: live text+reasoning chars beyond the last usage snapshot,
  // plus streaming tool args, ÷ 4 ≈ tokens — the estimate shows as soon as
  // the first token streams instead of waiting for the first usage event.
  const inFlightChars = Math.max(0, turnOutputChars - turnOutputCharsAtUsage) + turnArgChars;
  const inFlightTokens = Math.round(inFlightChars / 4);
  const liveTokens = turnTokens + inFlightTokens;
  const tok = liveTokens > 0 ? ' · ↓ ' + fmtTok(liveTokens) + ' tokens' : '';
  const outTok = turnOutputTokens + inFlightTokens;
  const modelElapsedMs = modelActiveMs + (modelActiveAt ? Date.now() - modelActiveAt : 0);
  const tps = outTok > 0 && modelElapsedMs >= 500 ? Math.round(outTok / (modelElapsedMs / 1000)) : null;
  const tpsStr = tps !== null ? ' · ' + tps + ' tokens/s' : '';
  return word + '… ' + fmtElapsed(ms) + tpsStr + tok;
}
function updateRunStrip() {
  const on = running || waitingPrompt !== null || retryStatus !== null;
  runStrip.style.display = on ? '' : 'none';
  runStrip.className = 'composer-run-strip' + (waitingPrompt !== null ? ' composer-run-strip--waiting' : '');
  // Desktop parity: the approval modebar stays usable while running, but is
  // disabled while a decision surface (approval/ask card) owns the footer.
  modebar.classList.toggle('composer-modebar--disabled', waitingPrompt !== null);
  if (!on) { runStripText.textContent = ''; runStripAnnounce.textContent = ''; return; }
  let stable = null;
  if (retryStatus) {
    stable = __('retrying_status').replace('{attempt}', retryStatus.attempt).replace('{max}', retryStatus.max);
  } else if (waitingPrompt === 'approval') {
    stable = pendingApprovalLabel === 'exit_plan_mode' ? __('run_waiting_plan') : __('run_waiting_approval').replace('{tool}', pendingApprovalLabel);
  } else if (waitingPrompt === 'ask') {
    stable = __('run_waiting_ask');
  }
  if (stable !== null) {
    runStripText.textContent = stable;
    runStripText.removeAttribute('aria-hidden');
    runStripAnnounce.textContent = stable;
  } else {
    runStripText.textContent = tickerText();
    runStripText.setAttribute('aria-hidden', 'true');
    runStripAnnounce.textContent = __('run_announce_running');
  }
}

function setRunning(on) {
  running=on;
  retryStatus=null;
  // While a turn runs the send button becomes the guidance-queue button
  // (desktop parity): Enter / click queue the draft instead of submitting.
  btnSend.classList.toggle('composer__btn--steer', on);
  btnSend.querySelector('svg').innerHTML = on
    ? '<path d="m15 10 5 5-5 5"/><path d="M4 4v7a4 4 0 0 0 4 4h12"/>'
    : '<path d="M12 19V5"/><path d="m5 12 7-7 7 7"/>';
  btnSend.title = on ? __('guidance_send') : __('send');
  btnSend.style.display='';
  btnStop.style.display=on?'':'none';
  input.placeholder = on ? __('placeholder_running') : (goalMode ? __('goal_placeholder') : __('placeholder'));
  statusDotSidebar.className=on?'status__dot status__dot--busy':'status__dot';
  setActionDisabled('btn-new', on, on?__('new_session_busy_title'):'');
  updateActionAvailability();
  if(on){turnStartAt=Date.now();turnTokens=0;turnOutputTokens=0;turnOutputChars=0;turnOutputCharsAtUsage=0;modelActiveAt=0;modelActiveMs=0;waitAccumMs=0;waitStartedAt=0;tickTimer=setInterval(updateRunStrip,1000);} else {clearInterval(tickTimer);}
  updateRunStrip();
}

function setRetrying(attempt,max) {
  retryStatus={attempt:attempt||0,max:max||0};
  statusDotSidebar.className='status__dot status__dot--busy';
  updateRunStrip();
}

function clearRetrying() {
  if(!retryStatus)return;
  retryStatus=null;
  updateRunStrip();
}

function setConnState(state) {
  // state: 'connected' | 'reconnecting' | 'disconnected'
  const colors={connected:'var(--success)',reconnecting:'var(--warning)',disconnected:'var(--danger)'};
  const labels={connected:__('connected'),reconnecting:__('reconnecting'),disconnected:__('disconnected')};
  statusDotSidebar.style.background=colors[state]||'';
  statusDotSidebar.className='status__dot'+(state==='reconnecting'?' status__dot--busy':'');
  if(!running)statusDotSidebar.title=labels[state]||state;
}

// ── transcript items model ──
// items is the single source of truth for the visible transcript. The SSE
// stream grows it incrementally; /history rebuilds it wholesale; both render
// through the same item→DOM functions, so a reload shows exactly what the
// live stream showed (and vice versa). An assistant item owns the tool calls
// it dispatched; a 'tool' item only appears when /history has a result whose
// dispatch is missing.
//   {id, kind:'user', text}
//   {id, kind:'assistant', text, reasoning, tools:[{id,name,args,output,err,status,diff,added,removed,readOnly,audit}], done, dom:{...}}
//   {id, kind:'tool', name, args, output, err, status}
//   {id, kind:'phase'|'notice'|'compaction'|'error', text, level?, detail?}
let items = [];
let deliveryRecoveryActive = false;
let nextItemId = 1;
let liveAssistant = null; // assistant item currently streaming
let currentTurn = 0; // incremented by turn_started; user messages claim turn+1
// While a history rebuild is in flight (initial load, /resume, /switch), the
// transcript is owned by the rebuild: SSE render events arriving in the gap
// would be wiped by the rebuild's innerHTML clear, so they are skipped. The
// rebuild's /history snapshot already contains everything that happened in
// the gap; fetchStatus afterwards re-syncs running state.
let historyPending = false;
const turnEls = new Map(); // turn number -> { el, tools: Map, summary }
function genItemId() { return 'it' + (nextItemId++); }
function resetItems() { items = []; nextItemId = 1; liveAssistant = null; deliveryRecoveryActive = false; currentTurn = 0; turnEls.clear(); if (jumpBarEl) { jumpBarEl.remove(); jumpBarEl = null; } jumpActiveId = null; jumpScrollPinned = true; if (scrollAnimFrame !== null) { cancelAnimationFrame(scrollAnimFrame); scrollAnimFrame = null; } }
function hasVisibleItems() {
  return items.some(it => it.kind === 'user' || it.kind === 'assistant' || it.kind === 'tool');
}
// textOut returns the plain text of an item for copying/export purposes.
function itemPlainText(it) {
  if (it.kind === 'user') return it.text || '';
  if (it.kind === 'assistant') {
    const parts = [];
    if (it.reasoning) parts.push(it.reasoning);
    if (it.text) parts.push(it.text);
    return parts.join('\n\n');
  }
  return '';
}
function findTool(toolId) {
  for (const it of items) {
    if (it.kind !== 'assistant' || !it.tools) continue;
    const t = it.tools.find(x => x.id === toolId);
    if (t) return { item: it, tool: t };
  }
  return null;
}
// legacy alias kept for callers that only need the state transition
function finalizeAssistant() {
  if (!liveAssistant) return;
  liveAssistant.done = true;
  liveAssistant = null;
}

// ── message actions (copy / edit) ──
let editingItem = null;
function addMsgActions(it, d, opts) {
  const actions = el('div', 'msg-actions');
  const copy = el('button', 'msg-action', '⧉');
  copy.type = 'button';
  copy.title = __('tool_copy');
  copy.onclick = async () => {
    const ok = await copyText(itemPlainText(it));
    copy.title = ok ? __('tool_copied') : __('tool_copy');
    setTimeout(() => { copy.title = __('tool_copy'); }, 1200);
  };
  actions.appendChild(copy);
  if (opts && opts.editable) {
    const edit = el('button', 'msg-action', '✎');
    edit.type = 'button';
    edit.title = __('edit_message');
    edit.onclick = () => startEdit(it, d);
    actions.appendChild(edit);
  }
  d.appendChild(actions);
}
function startEdit(it, d) {
  if (editingItem) return;
  editingItem = it;
  const textEl = d.querySelector('.msg__text');
  if (!textEl) { editingItem = null; return; }
  textEl.style.display = 'none';
  const box = el('div', 'msg-edit');
  const ta = document.createElement('textarea');
  ta.className = 'msg-edit__input';
  ta.value = it.text;
  ta.spellcheck = false;
  const bar = el('div', 'msg-edit__bar');
  const hint = el('span', 'msg-edit__hint', __('edit_hint'));
  const save = el('button', 'msg-edit__btn msg-edit__btn--save', __('edit_save'));
  const cancel = el('button', 'msg-edit__btn', __('edit_cancel'));
  save.type = 'button'; cancel.type = 'button';
  bar.appendChild(hint); bar.appendChild(save); bar.appendChild(cancel);
  box.appendChild(ta); box.appendChild(bar);
  d.appendChild(box);
  const finish = (saveIt) => {
    box.remove();
    textEl.style.display = '';
    editingItem = null;
    if (saveIt) {
      const v = ta.value;
      post('/edit', { display: v, input: v, original: it.text }).then(r => {
        if (r.ok) {
          it.text = v;
          syncQuestionNav();
          const sec = d.querySelector('.md-sections');
          if (sec) { sec.innerHTML = renderMarkdown(v); fixImageSrcs(sec); }
          let note = d.querySelector('.msg__edited');
          if (!note) {
            note = el('span', 'msg__edited', __('edited_note'));
            const wrap = d.querySelector('.msg__text');
            if (wrap && wrap.nextSibling) wrap.after(note);
            else d.appendChild(note);
          }
        }
      });
    }
  };
  save.onclick = () => finish(true);
  cancel.onclick = () => finish(false);
  ta.onkeydown = e => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); finish(true); }
    if (e.key === 'Escape') { e.preventDefault(); finish(false); }
  };
  ta.focus();
  ta.setSelectionRange(ta.value.length, ta.value.length);
}

// ── turn grouping ──
// Items render inside per-turn containers. Each turn carries a summary bar
// (first child) listing its tool calls; clicking the bar folds the turn's
// reasoning blocks and tool cards into it, leaving messages readable.
function turnContainer(turn) {
  let t = turnEls.get(turn);
  if (!t) {
    const container = document.createElement('div');
    container.className = 'turn';
    container.dataset.turn = turn;
    const summary = el('button', 'turn-summary');
    summary.type = 'button';
    summary.style.display = 'none';
    summary.onclick = () => toggleTurnFold(turn);
    container.appendChild(summary);
    log.appendChild(container);
    // userOverride: once the user manually folds/expands, auto fold/collapse
    // (turn_done, history rebuild) no longer applies — desktop behavior.
    t = { el: container, summary, tools: new Map(), folded: false, userOverride: false };
    turnEls.set(turn, t);
  }
  return t;
}
function registerTurnTool(t, tool) {
  t.tools.set(tool.id, tool);
}
// updateTurnSummaryForItem recomputes the turn's summary bar from the items
// model (source of truth) — tool cards attached to a streaming assistant
// never pass through appendItem, so the per-turn map alone would miss them.
function updateTurnSummaryForItem(it) {
  const t = turnEls.get(it.turn);
  if (!t) return;
  const tools = [];
  let thoughts = 0, running = false;
  for (const item of items) {
    if ((item.turn || 0) !== it.turn) continue;
    if (item.kind === 'tool') { tools.push(item); if (item.status === 'running') running = true; }
    else if (item.kind === 'assistant') {
      if (item.tools) {
        tools.push(...item.tools);
        if (item.tools.some(t => t.status === 'running')) running = true;
      }
      if (item.reasoning) thoughts++;
    }
  }
  updateTurnSummary(t, tools, thoughts, running);
}
// workLabel renders the desktop-style count copy: "Worked · 2 tools · 1
// thought" (or the running "Working…" while tools are still executing).
function workLabel(tools, thoughts, running) {
  if (running) return __('work_working');
  const zh = __LANG === 'zh';
  const t = tools + (zh ? ' 个工具' : (tools === 1 ? ' tool' : ' tools'));
  const th = thoughts + (zh ? ' 段思考' : (thoughts === 1 ? ' thought' : ' thoughts'));
  return __('work_done') + ' · ' + t + ' · ' + th;
}
function updateTurnSummary(t, tools, thoughts, running) {
  if (!tools.length) { t.summary.style.display = 'none'; return; }
  t.summary.textContent = workLabel(tools.length, thoughts, running);
  t.summary.style.display = '';
}
function applyTurnFold(t) {
  t.el.querySelectorAll('.reasoning, .card').forEach(n => {
    n.style.display = t.folded ? 'none' : '';
  });
  t.summary.classList.toggle('turn-summary--folded', !!t.folded);
}
function toggleTurnFold(turn) {
  const t = turnEls.get(turn);
  if (!t) return;
  t.userOverride = true;
  t.folded = !t.folded;
  applyTurnFold(t);
}
// appendItem appends a rendered item into its turn container and keeps the
// turn's summary bar up to date.
function appendItem(it, dom) {
  if (!dom) return;
  const t = turnContainer(it.turn || currentTurn);
  if (it.kind === 'user') {
    // Desktop order: user message first, then the fold summary, then
    // reasoning / tools / answer — insert the user message before the bar.
    t.el.insertBefore(dom, t.summary);
  } else {
    t.el.appendChild(dom);
  }
  if (t.folded && (dom.classList.contains('card') || dom.classList.contains('reasoning') || dom.querySelector('.card, .reasoning'))) {
    applyTurnFold(t);
  }
  updateTurnSummaryForItem(it);
  return t;
}

// ── item rendering ──
// The stream path mutates a live item's cached DOM in place (cheap, no
// re-layout of the whole transcript); renderItem builds a fresh DOM subtree
// for history rebuilds. Both read the same item state.
function renderItem(it) {
  if (it.kind === 'user') {
    const d = el('div', 'msg msg--user');
    d.id = 'question-anchor-' + it.id; // jump-bar scroll target (desktop parity)
    const wrap = el('div', 'msg__text msg__text--md');
    const sections = el('div', 'md-sections');
    const tail = el('span', 'md-tail');
    wrap.appendChild(sections); wrap.appendChild(tail);
    d.appendChild(wrap);
    if (it.text) {
      sections.innerHTML = renderMarkdown(it.text);
      fixImageSrcs(sections);
    }
    tail.style.display = 'none';
    addMsgActions(it, d, { editable: true });
    it.dom = { root: d, text: wrap, sections };
    return d;
  }
  if (it.kind === 'assistant') {
    const d = el('div', 'msg msg--assistant');
    d.dataset.done = it.done ? 'true' : 'false';
    it.dom = { root: d, textEl: null, reasoningEl: null, tools: new Map() };
    if (it.reasoning) appendAssistantReasoning(it, d);
    if (it.text) appendAssistantText(it, d);
    (it.tools || []).forEach(t => { const card = buildToolCard(t); if (card) { d.appendChild(card); it.dom.tools.set(t.id, card); } });
    if (it.done) { const c = d.querySelector('.cursor'); if (c) c.remove(); }
    addMsgActions(it, d, { editable: false });
    return d;
  }
  if (it.kind === 'phase') return el('div', 'phase', it.text || '');
  if (it.kind === 'notice') return el('div', 'notice' + (it.level === 'warn' ? ' notice--warn' : ''), (it.level === 'warn' ? '! ' : '') + (it.text || ''));
  if (it.kind === 'delivery') return renderDeliveryCard(it);
  if (it.kind === 'error') return el('div', 'msg--error', '✗ ' + (it.text || ''));
  if (it.kind === 'compaction') return renderCompactionItem(it);
  if (it.kind === 'tool') {
    const card = buildToolCard(it);
    if (card) { it.dom = { root: card }; }
    return card;
  }
  return null;
}
// ── question jump bar (desktop QuestionJumpBar parity) ──
// A thin rail on the transcript's right edge marks each user question.
// Hover ripples the nearest dots; clicking (or pressing the rail) smooth-
// scrolls to that question and detaches auto-pin, mirroring desktop.
const QUESTION_NAV_MIN_COUNT = 2;
let jumpBarEl = null;
let jumpActiveId = null; // last jumped / newest question id (accent dot)
let jumpHoverIdx = -1;   // hovered dot index, -1 when none
let jumpPreviewTop = 0;
let jumpScrollPinned = true; // rail follows the newest question; manual scroll-up detaches

function questionAnchors() {
  const anchors = [];
  let turn = 0;
  for (const it of items) {
    if (it.kind !== 'user') continue;
    anchors.push({ id: it.id, text: compactText(it.text, 80), turn });
    turn += 1;
  }
  return anchors;
}

function jumpQuestionById(id) {
  return questionAnchors().find((q) => q.id === id) || null;
}

function updateJumpDots(questions) {
  if (!jumpBarEl) return;
  const dots = jumpBarEl.querySelectorAll('.jump-item .jump-dot');
  dots.forEach((dot, idx) => {
    const q = questions[idx];
    const isActive = q && q.id === jumpActiveId;
    let width = isActive ? '18px' : '12px';
    let background = isActive ? 'var(--accent)' : '';
    let delay = '0ms';
    if (jumpHoverIdx >= 0) {
      const d = Math.abs(idx - jumpHoverIdx);
      width = d === 0 ? '32px' : d === 1 ? '20px' : d === 2 ? '14px' : width;
      // Desktop color ladder (.jump-dot[data-d="0/1/2"] in styles.css):
      // hovered dot = full accent, neighbors = 60% / 35% accent.
      if (d <= 2) {
        background = d === 0 ? 'var(--accent)'
          : d === 1 ? 'color-mix(in srgb, var(--accent) 60%, transparent)'
          : 'color-mix(in srgb, var(--accent) 35%, transparent)';
      }
      delay = (d * 20) + 'ms';
    }
    dot.style.width = width;
    dot.style.background = background;
    dot.style.transitionDelay = delay;
  });
}

function updateJumpPreview(questions, show, idx) {
  if (!jumpBarEl) return;
  const preview = jumpBarEl.querySelector('.jump-preview');
  if (!preview) return;
  if (!show || idx < 0 || idx >= questions.length) { preview.style.display = 'none'; return; }
  preview.style.top = jumpPreviewTop + 'px';
  const text = preview.querySelector('.jump-text');
  if (text) text.textContent = questions[idx].text;
  preview.style.display = '';
}

function closestJumpQuestion(questions, clientY) {
  if (!jumpBarEl) return null;
  const markers = jumpBarEl.querySelectorAll('.jump-item');
  const barRect = jumpBarEl.getBoundingClientRect();
  let closest = -1, closestDist = Infinity, closestY = 0;
  markers.forEach((item, index) => {
    const rect = item.getBoundingClientRect();
    const midY = rect.top + rect.height / 2;
    const dist = Math.abs(clientY - midY);
    if (dist < closestDist) { closestDist = dist; closest = index; closestY = midY - barRect.top; }
  });
  const question = questions[closest];
  if (!question) return null;
  return { question, index: closest, previewY: closestY };
}

// smoothScrollToTop animates #log's scrollTop with the same parameters the
// desktop uses (desktop/frontend/src/lib/gsapAnimations.ts: DUR_FAST*2 =
// 0.24s, EASE_OUT = "power2.out" ≈ cubic-bezier(0.2, 0.72, 0.2, 1)), so the
// two frontends share the same scroll feel without bundling gsap. Honors
// prefers-reduced-motion with an instant jump. A new call cancels any
// in-flight animation (desktop: gsap.killTweensOf). Scroll events fire on
// each frame, so pinnedToBottom stays in sync automatically.
let scrollAnimFrame = null;
function smoothScrollToTop(targetTop) {
  if (scrollAnimFrame !== null) { cancelAnimationFrame(scrollAnimFrame); scrollAnimFrame = null; }
  const el = log;
  const to = Math.max(0, targetTop);
  const from = el.scrollTop;
  if (Math.abs(to - from) < 0.5) return;
  if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
    el.scrollTop = to;
    return;
  }
  const DUR = 0.24; // desktop DUR_FAST (0.12) * 2
  const easeOut = (t) => 1 - Math.pow(1 - t, 2); // gsap "power2.out"
  const start = performance.now();
  const step = (now) => {
    const t = Math.min(1, (now - start) / (DUR * 1000));
    el.scrollTop = from + (to - from) * easeOut(t);
    if (t < 1) scrollAnimFrame = requestAnimationFrame(step);
    else scrollAnimFrame = null;
  };
  scrollAnimFrame = requestAnimationFrame(step);
}

function jumpToQuestion(q) {
  const node = document.getElementById('question-anchor-' + q.id);
  if (!node) return;
  pinnedToBottom = false; // detach auto-pin so the jump is not overridden
  jumpActiveId = q.id;
  // Unfold the target turn so its work is visible. Desktop equivalence: it
  // loads the turn's cold page (warmColdPageForTurn / warmLayerWithColdPageAtLeast)
  // and expands it (warmLayerWithExpandedTurn) before scrolling; webui renders
  // the whole history, so the only remaining "hidden work" state is a folded
  // turn — expanding it here keeps the jump semantics identical.
  const turnWrap = node.closest('.turn');
  if (turnWrap) {
    const t = turnEls.get(Number(turnWrap.dataset.turn));
    if (t && t.folded) { t.folded = false; applyTurnFold(t); }
  }
  const top = node.getBoundingClientRect().top - log.getBoundingClientRect().top + log.scrollTop - 12;
  smoothScrollToTop(top);
  updateJumpDots(questionAnchors());
  // Reveal the jumped dot if it sits outside the rail's 240px window.
  const idx = questionAnchors().findIndex((x) => x.id === q.id);
  if (idx >= 0) revealJumpItem(idx);
}

function revealJumpItem(idx) {
  const scroll = jumpBarEl && jumpBarEl.querySelector('.jump-scroll');
  if (!scroll) return;
  const item = scroll.children[idx];
  if (!item) return;
  const maxTop = scroll.scrollHeight - scroll.clientHeight;
  scroll.scrollTop = Math.max(0, Math.min(item.offsetTop - scroll.clientHeight / 2, maxTop));
}

function buildQuestionJumpBar(questions) {
  const nav = el('nav', 'jump-bar');
  nav.setAttribute('aria-label', __('question_nav_label'));
  const scroll = el('div', 'jump-scroll');
  questions.forEach((q, idx) => {
    const item = el('button', 'jump-item');
    item.type = 'button';
    item.dataset.id = q.id;
    item.setAttribute('aria-label', __('question_nav_jump').replace('{n}', String(idx + 1)));
    item.appendChild(el('span', 'jump-dot'));
    scroll.appendChild(item);
  });
  nav.appendChild(scroll);
  const preview = el('div', 'jump-preview');
  preview.style.display = 'none';
  preview.appendChild(el('span', 'jump-text'));
  nav.appendChild(preview);

  nav.addEventListener('mousemove', (e) => {
    const questions = questionAnchors();
    const closest = closestJumpQuestion(questions, e.clientY);
    if (!closest) return;
    jumpHoverIdx = closest.index;
    jumpPreviewTop = closest.previewY;
    updateJumpDots(questions);
    updateJumpPreview(questions, true, closest.index);
  });
  nav.addEventListener('mouseleave', () => {
    const questions = questionAnchors();
    jumpHoverIdx = -1;
    updateJumpDots(questions);
    updateJumpPreview(questions, false, -1);
  });
  scroll.addEventListener('scroll', () => {
    // Follow-the-newest pin: detached while the user scrolls up the rail,
    // re-attached once they return to the bottom (mirrors pinnedToBottom).
    jumpScrollPinned = scroll.scrollHeight - scroll.scrollTop - scroll.clientHeight < 8;
  });
  scroll.addEventListener('mousedown', (e) => {
    e.preventDefault();
    const questions = questionAnchors();
    const item = e.target.closest ? e.target.closest('.jump-item') : null;
    const q = item ? jumpQuestionById(item.dataset.id) : (closestJumpQuestion(questions, e.clientY) || {}).question;
    if (!q) return;
    jumpToQuestion(q);
  });
  scroll.addEventListener('click', (e) => {
    e.stopPropagation();
    const item = e.target.closest ? e.target.closest('.jump-item') : null;
    if (item && e.detail === 0) { // keyboard activation (Enter/Space)
      const q = jumpQuestionById(item.dataset.id);
      if (q) jumpToQuestion(q);
    }
  });
  return nav;
}

// syncQuestionNav rebuilds (or removes) the jump bar from the items model.
// Call after any change that alters the set/text of user messages.
function syncQuestionNav() {
  if (jumpBarEl) { jumpBarEl.remove(); jumpBarEl = null; }
  jumpHoverIdx = -1;
  const questions = questionAnchors();
  if (questions.length < QUESTION_NAV_MIN_COUNT) { jumpActiveId = null; return; }
  if (jumpActiveId === null || !questions.some((q) => q.id === jumpActiveId)) {
    jumpActiveId = questions[questions.length - 1].id;
  }
  jumpBarEl = buildQuestionJumpBar(questions);
  updateJumpDots(questions);
  // Mount on the non-scrolling .app (log's parent): #log itself scrolls, so
  // an absolutely-positioned bar inside it would ride the content and vanish
  // when scrolled down. Desktop mounts on the non-scrolling .transcript-shell.
  (log.parentElement || log).appendChild(jumpBarEl);
  // Start (and re-pin after a new question) at the newest entry; the user
  // can scroll the rail upward to reach older questions. Must run after the
  // mount: a detached element has no layout, so scrollHeight would be 0.
  if (jumpScrollPinned) {
    const scroll = jumpBarEl.querySelector('.jump-scroll');
    if (scroll) scroll.scrollTop = scroll.scrollHeight;
  }
}
function appendUserMsg(text) {
  const it = { id: genItemId(), kind: 'user', text, turn: currentTurn + 1 };
  items.push(it);
  hasVisibleHistory = true; updateActionAvailability(); hideWelcome();
  const d = renderItem(it); if (d) appendItem(it, d);
  scrollDown(true);
  syncQuestionNav();
  liveAssistant = null;
  return it;
}
function ensureAssistant() {
  if (liveAssistant) return liveAssistant;
  hasVisibleHistory = true; updateActionAvailability(); hideWelcome();
  const it = { id: genItemId(), kind: 'assistant', text: '', reasoning: '', tools: [], done: false, dom: null, turn: currentTurn };
  items.push(it);
  const d = renderItem(it); if (d) appendItem(it, d);
  liveAssistant = it;
  return it;
}
function setReasoningOpen(wrapper, open) {
  const body = wrapper.querySelector('.reasoning__body');
  const chevron = wrapper.querySelector('.reasoning__chevron');
  const summary = wrapper.querySelector('.reasoning__summary');
  if (!body) return;
  body.style.display = open ? '' : 'none';
  if (summary) summary.style.display = open ? 'none' : '';
  if (chevron) chevron.className = 'reasoning__chevron' + (open ? ' reasoning__chevron--open' : '');
}
// truncateReasoningText mirrors desktop's displayReasoningText streaming
// truncation: keep the tail, cap at 12k chars / 240 lines. Only applied to
// live appends; history rebuilds show the full reasoning.
// reasoningSummaryText mirrors desktop's lib/reasoningSummary.ts: while
// streaming the tail line is the live signal, once settled the head line
// reads best. Single line, whitespace collapsed, code-point-safe, 180 chars.
function reasoningSummaryText(text, streaming) {
  const MAX_CHARS = 180;
  const s = String(text);
  if (!s) return '';
  let line = '';
  if (streaming) {
    let end = s.length;
    while (end > 0) {
      while (end > 0 && (s[end-1]==='\n' || s[end-1]==='\r')) end--;
      let start = end;
      while (start > 0 && s[start-1]!=='\n' && s[start-1]!=='\r') start--;
      const t = s.slice(start, end).trim();
      if (t) { line = t; break; }
      end = start;
    }
  } else {
    let i = 0, n = s.length;
    while (i < n) {
      while (i < n && (s[i]==='\n' || s[i]==='\r')) i++;
      const start = i;
      while (i < n && s[i]!=='\n' && s[i]!=='\r') i++;
      const t = s.slice(start, i).trim();
      if (t) { line = t; break; }
    }
  }
  if (!line) return '';
  const normalized = line.replace(/\s+/g, ' ');
  const chars = Array.from(normalized);
  if (chars.length <= MAX_CHARS) return normalized;
  return MAX_CHARS <= 1 ? '…' : chars.slice(0, MAX_CHARS - 1).join('') + '…';
}
function truncateReasoningText(text) {
  const MAX_CHARS = 12000, MAX_LINES = 240;
  const s = String(text);
  const lines = s.split('\n');
  // Line cap first: a wide-but-short reasoning (many lines, few chars) must
  // still truncate; then the character cap for single-line monsters.
  if (lines.length > MAX_LINES) return '...\n' + lines.slice(lines.length - MAX_LINES).join('\n');
  if (s.length > MAX_CHARS) return '...\n' + s.slice(s.length - MAX_CHARS);
  return s;
}
// appendAssistantReasoning inserts the reasoning block before the text block
// (or appends it) and streams new chunks into it. Desktop behavior: opens
// while streaming, auto-collapses on completion, and a manual toggle wins
// over both (reasoningUserOverride). The head shows "Thinking…" while
// running and "Thinking done · Ns" after, with a shimmer on the live label.
function appendAssistantReasoning(it, container) {
  const w = el('div', 'reasoning');
  const b = el('button', 'reasoning__head');
  b.type = 'button';
  b.innerHTML = '<span class="reasoning__chevron">▶</span><span class="reasoning__label">' + __('thinking') + '</span>';
  const summary = el('button', 'reasoning__summary');
  summary.type = 'button';
  const body = el('div', 'reasoning__body');
  body.style.display = 'none';
  const toggle = () => {
    it.reasoningUserOverride = true; // manual toggle wins over auto collapse/expand
    setReasoningOpen(w, body.style.display === 'none');
  };
  b.onclick = toggle; summary.onclick = toggle;
  w.appendChild(b); w.appendChild(summary); w.appendChild(body);
  if (it.reasoning) body.textContent = it.done ? it.reasoning : truncateReasoningText(it.reasoning);
  summary.textContent = reasoningSummaryText(it.reasoning, !it.done);
  if (!it.dom) it.dom = { root: null, textEl: null, reasoningEl: null, tools: new Map() };
  const textEl = it.dom.textEl && it.dom.textEl.parentNode;
  if (textEl) container.insertBefore(w, textEl); else container.appendChild(w);
  it.dom.reasoningEl = body;
  if (!it.done) it.reasoningStartedAt = Date.now();
  else b.querySelector('.reasoning__label').textContent = __('thinking_done'); // history rebuild: settled label, no duration
  w.dataset.running = it.done ? 'false' : 'true';
  // Desktop parity: thinking renders collapsed with a horizontal summary by
  // default (live turns included); clicking the head or summary expands the
  // vertical trace.
  let display='closed';try{display=localStorage.getItem('baize-reasoning-display')||'closed';}catch{}
  setReasoningOpen(w,display==='open'||(display==='auto'&&!it.done));
  return w;
}
function appendReasoning(t) {
  const it = ensureAssistant();
  it.reasoning += t;
  if (!it.dom.reasoningEl) appendAssistantReasoning(it, it.dom.root); // seeds existing content
  else it.dom.reasoningEl.textContent = truncateReasoningText(it.reasoning);
  const w = it.dom.reasoningEl.parentNode;
  const summary = w && w.querySelector('.reasoning__summary');
  if (summary && summary.style.display !== 'none') {
    summary.textContent = reasoningSummaryText(it.reasoning, true); // tail line while streaming
    summary.scrollLeft = summary.scrollWidth; // follow the newest text
  }
  scrollDown();
}
// ── streaming markdown rendering ──
// Assistant text streams through a throttled markdown pipeline: the text is
// split at paragraph boundaries (\n\n). Paragraphs that are complete render
// once as sanitized HTML (they never change again); the incomplete tail
// paragraph stays a plain-text span that receives chunks directly — no
// re-render, no flicker. A fenced code block whose opening fence is still in
// the stable part keeps its paragraphs in the tail until fences balance, so
// the renderer never sees an unterminated block.
const MD_THROTTLE_MS = 80;
function countFences(s) {
  let n = 0;
  for (const line of String(s).split('\n')) {
    if (/^\s*(`{3,}|~{3,})/.test(line)) n++;
  }
  return n;
}
function splitMd(text) {
  const idx = String(text).lastIndexOf('\n\n');
  if (idx < 0) return { stable: '', tail: String(text) };
  let stable = String(text).slice(0, idx);
  let tail = String(text).slice(idx + 2);
  // Pull paragraphs back into the tail while the stable part ends inside an
  // open fenced block (rare but breaks the renderer if left).
  for (let guard = 0; guard < 8 && countFences(stable) % 2 === 1; guard++) {
    const j = stable.lastIndexOf('\n\n');
    if (j < 0) { tail = stable + '\n\n' + tail; stable = ''; break; }
    tail = stable.slice(j + 2) + '\n\n' + tail;
    stable = stable.slice(0, j);
  }
  return { stable, tail };
}
// Links open in a new tab and never inherit opener access; DOMPurify already
// strips javascript:/data:text/html URLs and event handlers. DOMPurify 3.x
// registers hooks globally (config-level callbacks were removed).
DOMPurify.addHook('afterSanitizeAttributes', (node) => {
  if (node.tagName === 'A') {
    node.setAttribute('target', '_blank');
    node.setAttribute('rel', 'noopener noreferrer');
  }
});
function sanitizeMarkdownHtml(raw) {
  return DOMPurify.sanitize(raw, { USE_PROFILES: { html: true } });
}
function renderMarkdown(src) {
  const text = String(src || '');
  if (!text.trim()) return '';
  try {
    const raw = marked.parse(text, { gfm: true });
    return sanitizeMarkdownHtml(raw);
  } catch {
    return escHtml(text);
  }
}
function highlightBlocks(root) {
  root.querySelectorAll('pre code').forEach(block => {
    const m = (block.className || '').match(/language-([\w-]+)/);
    let html = '';
    try {
      if (m && hljs.getLanguage(m[1])) html = hljs.highlight(block.textContent, { language: m[1] }).value;
      else html = hljs.highlightAuto(block.textContent).value;
    } catch { html = escHtml(block.textContent); }
    block.innerHTML = html;
    const pre = block.parentElement;
    pre.classList.add('hljs');
    // Language label + copy button on a slim header bar.
    const head = el('div', 'code-head');
    head.appendChild(el('span', 'code-lang', m ? m[1] : 'code'));
    const copy = el('button', 'code-copy', __('tool_copy'));
    copy.type = 'button';
    copy.title = __('tool_copy');
    copy.onclick = async () => {
      const ok = await copyText(block.textContent);
      copy.textContent = ok ? __('tool_copied') : __('tool_copy');
      copy.title = ok ? __('tool_copied') : __('tool_copy');
      setTimeout(() => { copy.textContent = __('tool_copy'); copy.title = __('tool_copy'); }, 1200);
    };
    head.appendChild(copy);
    pre.prepend(head);
  });
}
// mdState per assistant item: { timer, renderedLen, sectionsEl, tailEl }
function mdState(it) {
  if (!it.md) it.md = { timer: null, renderedLen: 0, sectionsEl: null, tailEl: null };
  return it.md;
}
function renderMdNow(it, container) {
  const md = mdState(it);
  let stable, tail;
  if (it.done) {
    // Finalized: render everything, collapse the tail.
    stable = it.text;
    tail = '';
  } else {
    ({ stable, tail } = splitMd(it.text));
  }
  // Incremental: completed paragraphs never change, so only the newly
  // completed chunk needs rendering (avoids O(n²) re-renders on long turns).
  // splitMd keeps fences balanced, and both old and new stable prefixes are
  // balanced, so the chunk between them is too.
  if (stable.length > md.renderedLen) {
    const chunk = stable.slice(md.renderedLen);
    const tmp = document.createElement('div');
    tmp.innerHTML = renderMarkdown(chunk);
    fixImageSrcs(tmp);
    highlightBlocks(tmp);
    while (tmp.firstChild) md.sectionsEl.appendChild(tmp.firstChild);
    md.renderedLen = stable.length;
  }
  md.tailEl.textContent = tail;
  if (it.done && !tail) {
    // All paragraphs finalized: the tail span is empty and can go away.
    md.tailEl.style.display = 'none';
  }
}
// fixImageSrcs rewrites workspace-relative image references to the /file
// endpoint (the browser cannot read arbitrary local paths) and wires click
// handlers for the lightbox viewer.
function fixImageSrcs(root, basePath='') {
  root.querySelectorAll('img').forEach(img => {
    const src = img.getAttribute('src') || '';
    if (!src || /^(https?:|data:)/i.test(src)) return;
    const path = basePath ? workspaceRelativeFrom(basePath, src) : src;
    const url = '/file?path=' + encodeURIComponent(path);
    img.setAttribute('src', url);
    img.classList.add('md-image');
    img.addEventListener('click', e => { e.preventDefault(); openImageViewer(url); });
  });
  fixWorkspaceLinks(root);
}
function workspaceRelativeFrom(basePath, target) {
  const raw=String(target||'').replace(/\\/g,'/');
  if(!basePath||raw.startsWith('/')||/^[a-z]:\//i.test(raw))return raw;
  const parts=String(basePath).replace(/\\/g,'/').split('/');parts.pop();
  for(const part of raw.split('/')){if(!part||part==='.')continue;if(part==='..')parts.pop();else parts.push(part);}
  return parts.join('/');
}
function workspacePathCandidate(raw, allowBareName=false) {
  const value=String(raw||'').trim();
  if(!value||value.length>500||/[\r\n]/.test(value)||/^(https?:|mailto:|tel:|#|javascript:|data:)/i.test(value))return '';
  if(value.startsWith('/workspace/')||value.startsWith('/file?'))return '';
  if(/[\\/]/.test(value)||/^[a-z]:[\\/]/i.test(value)||(allowBareName&&/^[\w@(). -]+\.[a-z0-9]{1,10}$/i.test(value)))return value.replace(/^file:\/\//i,'');
  return '';
}
function fixWorkspaceLinks(root) {
  root.querySelectorAll('a[href]').forEach(anchor=>{
    const path=workspacePathCandidate(anchor.getAttribute('href'),true);
    if(!path)return;
    anchor.removeAttribute('target');anchor.removeAttribute('rel');anchor.dataset.workspacePath=path;
    anchor.addEventListener('click',event=>{event.preventDefault();openWorkspacePath(path);});
  });
  root.querySelectorAll('code').forEach(code=>{
    if(code.closest('pre')||code.closest('a'))return;
    const path=workspacePathCandidate(code.textContent,false);
    if(!path)return;
    code.classList.add('workspace-path-link');code.tabIndex=0;code.setAttribute('role','link');code.title=__('workspace_files');
    const open=event=>{if(event.type==='keydown'&&event.key!=='Enter'&&event.key!==' ')return;event.preventDefault();openWorkspacePath(path);};
    code.addEventListener('click',open);code.addEventListener('keydown',open);
  });
}
function scheduleMdRender(it) {
  const md = mdState(it);
  if (md.timer) return;
  md.timer = setTimeout(() => { md.timer = null; if (it.dom) renderMdNow(it, it.dom.root); }, MD_THROTTLE_MS);
}
// appendAssistantText attaches the message container: a sections div (rendered
// paragraphs) + a plain-text tail span (the paragraph still streaming).
function appendAssistantText(it, container) {
  if (!it.dom) it.dom = { root: container, textEl: null, reasoningEl: null, tools: new Map() };
  const wrap = el('div', 'msg__text msg__text--md');
  const md = mdState(it);
  md.sectionsEl = el('div', 'md-sections');
  md.tailEl = el('span', 'md-tail');
  wrap.appendChild(md.sectionsEl);
  wrap.appendChild(md.tailEl);
  container.appendChild(wrap);
  container.appendChild(el('span', 'cursor'));
  it.dom.textEl = wrap;
  renderMdNow(it, container);
}
function appendText(t) {
  const it = ensureAssistant();
  it.text += t;
  if (!it.dom.textEl) appendAssistantText(it, it.dom.root); // renders existing content
  else scheduleMdRender(it);
  scrollDown();
}
function finalizeMsg() {
  const it = liveAssistant;
  if (it) {
    it.done = true; // mark before the flush so the whole text renders and the tail collapses
    liveAssistant = null;
    if (it.md && it.md.timer) { clearTimeout(it.md.timer); it.md.timer = null; }
    if (it.dom && it.dom.textEl) renderMdNow(it, it.dom.root);
    if (it.dom && it.dom.root) {
      const r = it.dom.root.querySelector('.reasoning');
      if (r) {
        r.dataset.running = 'false';
        // "Thinking done · Ns" duration label (desktop ReasoningPanel).
        const label = r.querySelector('.reasoning__label');
        if (label && it.reasoningStartedAt) {
          label.textContent = __('thinking_done') + ' · ' + fmtElapsed(Date.now() - it.reasoningStartedAt);
        }
        // Auto-collapse only when the user hasn't manually toggled the block.
        if (!it.reasoningUserOverride) {let display='closed';try{display=localStorage.getItem('baize-reasoning-display')||'closed';}catch{}setReasoningOpen(r,display==='open');}
        const sm = r.querySelector('.reasoning__summary');
        if (sm) sm.textContent = reasoningSummaryText(it.reasoning, false); // head line once settled
      }
      const c = it.dom.root.querySelector('.cursor');
      if (c) c.remove();
    }
  }
}

// ── tool cards (item-backed) ──
function toolIcon(kind) {
  if(kind==='success')return '<svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>';
  if(kind==='danger')return '<svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>';
  return '<svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10" stroke-dasharray="60" stroke-dashoffset="20"/></svg>';
}
function toolArgsSummary(args) {
  const raw=String(args||'').trim();
  if(!raw)return '';
  try{
    const parsed=JSON.parse(raw);
    if(parsed&&typeof parsed==='object'&&!Array.isArray(parsed)){
      const key=['cmd','command','path','file','query','q','url','prompt','input'].find(k=>parsed[k]);
      if(key)return compactText(parsed[key],90);
    }
  }catch{}
  return compactText(raw,90);
}
function updateToolMeta(card, text) {
  const meta=card.querySelector('.card-meta');
  if(meta)meta.textContent=text;
}
function setToolStatus(card, tone, label) {
  card.dataset.tone=tone;
  const badge=card.querySelector('.card-badge');
  if(badge)badge.textContent=label;
  const ico=card.querySelector('.ico');
  if(ico){ico.className='ico'+(tone==='accent'?' spin':'');ico.innerHTML=toolIcon(tone);}
}
function copyText(text) {
  if(!text)return Promise.resolve(false);
  if(navigator.clipboard?.writeText)return navigator.clipboard.writeText(text).then(()=>true).catch(()=>false);
  return Promise.resolve(false);
}
// renderDiffView renders a unified-diff string as red/green lines with a
// +N -M stat recorded on the container (used for the collapsed card meta).
function renderDiffView(diffText) {
  const wrap = el('div', 'diff-view');
  const stats = { add: 0, del: 0 };
  String(diffText || '').split('\n').forEach(line => {
    let type = 'ctx';
    if (line.startsWith('+++') || line.startsWith('---') || line.startsWith('@@')) type = 'meta';
    else if (line.startsWith('+')) { type = 'add'; stats.add++; }
    else if (line.startsWith('-')) { type = 'del'; stats.del++; }
    const row = el('div', 'diff-line diff-line--' + type);
    row.textContent = line;
    wrap.appendChild(row);
  });
  wrap.dataset.stats = '+' + stats.add + ' -' + stats.del;
  return wrap;
}
// toolResultMeta renders the finished-state card meta by tool kind: errors
// show the duration only; diff-carrying tools (file edits) show the +/- line
// tallies; command tools show the duration only; anything else shows duration
// plus output line count (the original rule).
function toolResultMeta(t, durMs) {
  const dur = durMs>0 ? fmtElapsed(durMs) : '';
  const name = String(t.name||'').toLowerCase();
  const isCmd = name==='bash'||name==='powershell'||name==='cmd'||name==='shell'||name==='terminal';
  if (t.err) return dur;
  if (t.added||t.removed||t.diff) return '+'+(t.added||0)+' -'+(t.removed||0)+(dur?' · '+dur:'');
  if (isCmd) return dur;
  return (dur?dur+' · ':'') + lineCount(String(t.output||'')) + ' ' + __('tool_lines');
}
const SUBAGENT_PROGRESS_PREFIX='reasonix.subagent.';
const SUBAGENT_PROGRESS_LIMITS={reasoning:8<<10,text:8<<10,notice:2<<10};
const SUBAGENT_PHASES=new Set(['queued','running','reasoning','responding','tool','retrying','completed','failed','cancelled']);
function isSubagentTool(tool){
  if(!tool)return false;
  if(tool.subagent)return true;
  if(tool.profile&&typeof tool.profile==='object')return true;
  return ['task','read_only_task','parallel_tasks','fleet'].includes(String(tool.name||''));
}
function terminalSubagentPhase(phase){return phase==='completed'||phase==='failed'||phase==='cancelled';}
function subagentPhaseLabel(phase){return __("subagent_phase_"+(phase||'running'));}
function appendSubagentPreview(current,chunk,limit){
  const value=String(current||'')+String(chunk||'');
  if(typeof TextEncoder==='undefined')return {value:value.slice(-limit),truncated:value.length>limit};
  const encoder=new TextEncoder();if(encoder.encode(value).length<=limit)return {value,truncated:false};
  const chars=Array.from(value);let lo=0,hi=chars.length;
  while(lo<hi){const mid=Math.floor((lo+hi)/2);if(encoder.encode(chars.slice(mid).join('')).length<=limit)hi=mid;else lo=mid+1;}
  return {value:chars.slice(lo).join(''),truncated:true};
}
function ensureSubagentState(tool){
  if(!tool.subagent)tool.subagent={phase:'running',reasoning:'',text:'',notice:'',truncated:false,startedAt:Date.now(),lastActivityAt:Date.now(),finalOutput:''};
  return tool.subagent;
}
function subagentPreviewMode(){try{return localStorage.getItem('baize-subagent-preview')||'full';}catch{return 'full';}}
function subagentAutoCollapse(){try{return localStorage.getItem('baize-subagent-auto-collapse')!=='false';}catch{return true;}}
function appendSubagentSection(body,label,value,cls){
  if(!value)return;
  const section=el('section','subagent-preview__section '+cls);
  section.appendChild(el('div','subagent-preview__label',label));
  section.appendChild(el('pre','subagent-preview__text',value));
  body.appendChild(section);
}
function renderSubagentCard(card,tool){
  if(!card||!isSubagentTool(tool))return;
  const state=ensureSubagentState(tool);
  card.classList.add('card--subagent');
  const terminal=terminalSubagentPhase(state.phase);
  const tone=state.phase==='failed'?'danger':state.phase==='cancelled'?'danger':terminal?'success':'accent';
  setToolStatus(card,tone,subagentPhaseLabel(state.phase));
  const elapsed=state.durationMs||tool.durationMs||Math.max(0,Date.now()-state.startedAt);
  const profile=tool.profile||tool.subagentSummary||{};
  const profileText=[profile.model,profile.effort].filter(Boolean).join(' · ');
  updateToolMeta(card,[profileText,fmtElapsed(elapsed)].filter(Boolean).join(' · '));
  const body=card.querySelector('.card-body');
  if(!body)return;
  body.classList.add('subagent-preview');
  body.innerHTML='';
  if(subagentPreviewMode()==='compact'){
    body.appendChild(el('div','subagent-preview__activity',subagentPhaseLabel(state.phase)));
  }else{
    appendSubagentSection(body,__('subagent_reasoning'),state.reasoning,'subagent-preview__section--reasoning');
    appendSubagentSection(body,__('subagent_response'),state.text,'subagent-preview__section--text');
    appendSubagentSection(body,__('subagent_notice'),state.notice,'subagent-preview__section--notice');
    appendSubagentSection(body,__('subagent_result'),state.finalOutput,'subagent-preview__section--result');
    if(!state.reasoning&&!state.text&&!state.notice&&!state.finalOutput)body.appendChild(el('div','subagent-preview__activity',subagentPhaseLabel(state.phase)));
    if(state.truncated)body.appendChild(el('div','subagent-preview__truncated',__('subagent_truncated')));
  }
  if(card.dataset.open==='true')body.style.display='';
}
function updateSubagentElapsed(){
  items.forEach(it=>{
    const tools=it.kind==='assistant'?(it.tools||[]):(it.kind==='tool'?[it]:[]);
    tools.forEach(tool=>{if(isSubagentTool(tool)&&tool.subagent&&!terminalSubagentPhase(tool.subagent.phase)){const found=findToolItem(tool.id);if(found){const card=cardForTool(found.item,found.tool);if(card)renderSubagentCard(card,tool);}}});
  });
}
setInterval(updateSubagentElapsed,1000);
function renderToolAudit(card, tool) {
  let audit=card.querySelector('.card-audit');
  const entries=Array.isArray(tool.audit)?tool.audit:[];
  if(!entries.length){if(audit)audit.remove();return;}
  if(!audit){audit=el('div','card-audit');card.appendChild(audit);}
  audit.innerHTML='';
  audit.appendChild(el('div','card-audit__label',__('tool_audit')));
  entries.forEach(entry=>audit.appendChild(el('div','card-audit__line',entry.text||'')));
}
function addToolAudit(tool, entry) {
  if(!tool||!entry||!entry.text)return;
  if(!Array.isArray(tool.audit))tool.audit=[];
  if(!tool.audit.some(existing=>existing.text===entry.text&&existing.code===entry.code&&existing.detail===entry.detail))tool.audit.push(entry);
}
function capabilityAuditEntry(tool) {
  if(!tool.resolvedName||tool.resolvedName===tool.name)return null;
  return {text:'capability proxy: '+tool.name+' → '+tool.resolvedName,code:'capability_proxy',detail:String(tool.capabilityId||'')};
}
// buildToolCard renders a tool call item. A fresh card shows running state;
// finished tools get their status + output from the item fields.
function buildToolCard(tool) {
  const card=el('div','card');
  card.id='tool-'+tool.id;
  card.dataset.open=isSubagentTool(tool)&&tool.status!=='done'?'true':'false';
  const tone = tool.err ? 'danger' : (tool.status === 'done' || tool.status === 'running' ? (tool.status === 'done' ? 'success' : 'accent') : 'accent');
  card.dataset.tone = tone;
  card.dataset.toolArgs=String(tool.args||'');
  card.dataset.startedAt=String(Number(tool.startedAt||Date.now()));
  const head=el('div','card-head');
  const summary=toolArgsSummary(tool.args);
  head.innerHTML='<span class="ico'+(tone==='accent'?' spin':'')+'">'+toolIcon(tone==='danger'?'danger':tone==='success'?'success':'accent')+'</span><div class="card-main"><div class="card-title"><span class="name">'+escHtml(tool.name||'tool')+'</span>'+(summary?'<span class="subject">'+escHtml(summary)+'</span>':'')+'</div><div class="card-meta"></div></div><span class="card-badge">'+escHtml(tool.err?__('tool_failed'):tool.status==='done'?__('tool_done'):__('tool_running'))+'</span><div class="card-actions"><button type="button" class="card-action card-copy" title="'+escAttr(__('tool_copy'))+'" aria-label="'+escAttr(__('tool_copy'))+'"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg></button><span class="chev"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg></span></div>';
  const body=el('div','card-body');
  body.style.display='none';
  if (tool.diff) {
    const dv = renderDiffView(tool.diff);
    body.appendChild(dv);
    card.dataset.hasDiff = 'true';
    // The collapsed meta shows the stat; expanded body shows the lines.
    const meta = head.querySelector('.card-meta');
    if (meta) meta.textContent = dv.dataset.stats;
  }
  head.onclick=e=>{if(e.target.closest('button'))return;const open=card.dataset.open==='true';card.dataset.userOverride='true';card.dataset.open=open?'false':'true';body.style.display=open?'none':'';};
  const copy=head.querySelector('.card-copy');
  if(copy)copy.onclick=async e=>{e.stopPropagation();const audit=card.querySelector('.card-audit');const detail=[body.textContent,audit?.textContent].filter(Boolean).join('\n');const ok=await copyText(detail||card.dataset.toolArgs||'');copy.title=ok?__('tool_copied'):__('tool_copy');setTimeout(()=>{copy.title=__('tool_copy');},1200);};
  card.appendChild(head);card.appendChild(body);
  renderToolAudit(card,tool);
  if(!tool.diff&&(tool.output||tool.err)){body.textContent=(tool.output||'')+(tool.err?('\n'+tool.err):'');}
  // History rebuilds create cards for tools that already finished; render the
  // result meta here the same way renderToolResult would for a live stream.
  if (tool.status==='done' || tool.err) {
    const meta = head.querySelector('.card-meta');
    if (meta) meta.textContent = toolResultMeta(tool, tool.durationMs);
  }
  if(isSubagentTool(tool))renderSubagentCard(card,tool);
  return card;
}
function cardForTool(it, tool) {
  if (it.kind === 'assistant') {
    if (!it.dom || !it.dom.root) return null;
    let card = it.dom.tools.get(tool.id);
    if (!card) {
      card = buildToolCard(tool);
      let parentRoot=null;
      if(tool.parentId){
        const parent=findToolItem(tool.parentId);
        if(parent){const parentCard=cardForTool(parent.item,parent.tool);if(parentCard){let nest=parentCard.querySelector(':scope > .card-nest');if(!nest){nest=el('div','card-nest');nest.appendChild(el('div','card-nest__label',__('subagent_tools')));parentCard.appendChild(nest);}parentRoot=nest;}}
      }
      (parentRoot||it.dom.root).appendChild(card);
      if(tool.parentId&&!parentRoot)card.classList.add('card--subagent-orphan');
      it.dom.tools.set(tool.id, card);
    }
    return card;
  }
  // Standalone tool item: build lazily on first touch.
  if (!it.dom) { const card = buildToolCard(tool); it.dom = { root: card }; appendItem(it, card); }
  return it.dom.root;
}
function findToolItem(toolId) {
  const f = findTool(toolId);
  if (f) return f;
  const it = items.find(x => x.kind === 'tool' && x.id === toolId);
  return it ? { item: it, tool: it } : null;
}
function toolsInCurrentTurnNewestFirst() {
  const out=[];
  for(let i=items.length-1;i>=0;i--){
    const item=items[i];
    if((item.turn||0)!==currentTurn)continue;
    if(item.kind==='tool')out.push(item);
    else if(item.kind==='assistant'&&Array.isArray(item.tools)){
      for(let j=item.tools.length-1;j>=0;j--)out.push(item.tools[j]);
    }
  }
  return out;
}
function auditToolForNotice(e) {
  const tools=toolsInCurrentTurnNewestFirst();
  if(e.code==='capability_proxy'||String(e.text||'').startsWith('capability proxy:')){
    const capabilityId=String(e.detail||'');
    return tools.find(tool=>{
      if(String(tool.name||'')!=='use_capability'||tool.status!=='running')return false;
      if(String(tool.capabilityId||'')===capabilityId)return true;
      try{return String(JSON.parse(String(tool.args||'{}')).capability_id||'')===capabilityId;}catch(_){return false;}
    })||null;
  }
  const receipt=e.decisionReceipt;
  if(e.code!=='decision_receipt'||!receipt)return null;
  if(receipt.kind==='ask')return tools.find(tool=>String(tool.name||'')==='ask'&&tool.status==='running')||null;
  if(receipt.kind==='tool'&&receipt.tool)return tools.find(tool=>String(tool.name||'')===String(receipt.tool)&&tool.status==='running')||null;
  return null;
}
function attachAuditNotice(e) {
  const tool=auditToolForNotice(e);
  if(!tool)return false;
  const entry={text:String(e.text||''),code:String(e.code||''),detail:String(e.detail||'')};
  addToolAudit(tool,entry);
  const found=findToolItem(tool.id);
  if(found){const card=cardForTool(found.item,found.tool);if(card)renderToolAudit(card,found.tool);}
  return true;
}
function renderToolDispatch(tool) {
  if(hiddenTranscriptTool(tool&&tool.name))return;
  if(!tool.id)return;
  hideWelcome();
  let it, t;
  const found = findToolItem(tool.id);
  if (found) { it = found.item; t = found.tool; }
  else if (liveAssistant) {
    // Attach to the streaming assistant; a refreshed dependent writer shares
    // the same id and is merged above instead of duplicating a card.
    t = { id: tool.id, name: tool.name || 'tool', args: '', output: '', err: '', status: 'running', readOnly: !!tool.readOnly, parentId: tool.parentId || '', profile: tool.profile || null, startedAt: Number(tool.startedAt||Date.now()) };
    liveAssistant.tools.push(t);
    it = liveAssistant;
  } else {
    // No streaming assistant yet (tools can dispatch before text): a
    // standalone tool item that later turns group with its assistant.
    it = { id: tool.id, kind: 'tool', name: tool.name || 'tool', args: '', output: '', err: '', status: 'running', turn: currentTurn, parentId: tool.parentId || '', profile: tool.profile || null, startedAt: Number(tool.startedAt||Date.now()) };
    items.push(it);
    t = it;
  }
  if (tool.name) t.name = tool.name;
  if (tool.parentId) t.parentId = tool.parentId;
  if (tool.profile) t.profile = tool.profile;
  if (tool.startedAt) t.startedAt = Number(tool.startedAt);
  if(isSubagentTool(t))ensureSubagentState(t);
  if (tool.args) t.args = String(tool.args);
  if (tool.resolvedName) t.resolvedName = tool.resolvedName;
  if (tool.capabilityId) t.capabilityId = tool.capabilityId;
  addToolAudit(t,capabilityAuditEntry(t));
  if (tool.diff) t.diff = tool.diff;
  if (tool.added) t.added = tool.added;
  if (tool.removed) t.removed = tool.removed;
  const card = cardForTool(it, t);
  if (card) {
    const summary = toolArgsSummary(t.args);
    const title = card.querySelector('.card-title');
    if (title) title.innerHTML = '<span class="name">'+escHtml(t.name)+'</span>'+(summary?'<span class="subject">'+escHtml(summary)+'</span>':'');
    card.dataset.toolArgs = String(t.args || '');
    // A full dispatch may carry the diff after a partial one created the card.
    if (tool.diff && !card.dataset.hasDiff) {
      const body = card.querySelector('.card-body');
      if (body) {
        const dv = renderDiffView(tool.diff);
        body.appendChild(dv);
        card.dataset.hasDiff = 'true';
      }
    }
    renderToolAudit(card,t);
    if(isSubagentTool(t))renderSubagentCard(card,t);
  }
  if(t.parentId){const parent=findToolItem(t.parentId);if(parent&&isSubagentTool(parent.tool)){const state=ensureSubagentState(parent.tool);if(!terminalSubagentPhase(state.phase)){state.phase='tool';state.lastActivityAt=Date.now();const parentCard=cardForTool(parent.item,parent.tool);if(parentCard)renderSubagentCard(parentCard,parent.tool);}}}
  updateTurnSummaryForItem(it);
  scrollDown();
}
function shellDisplayName(ex){
  if(!ex||!ex.shell)return 'bash';
  if(ex.shell==='git-bash')return 'Git Bash';
  if(ex.shell==='powershell')return 'Windows PowerShell';
  if(ex.shell==='pwsh')return 'PowerShell 7+';
  return ex.shell;
}
function renderToolResult(tool) {
  if(hiddenTranscriptTool(tool&&tool.name))return;
  const found = findToolItem(tool.id);
  if (!found) return;
  const { item: it, tool: t } = found;
  t.err = tool.err || '';
  t.output = tool.output || '';
  t.status = tool.err ? 'failed' : 'done';
  t.durationMs = tool.durationMs;
  if(tool.startedAt)t.startedAt=Number(tool.startedAt);
  if(tool.endedAt)t.endedAt=Number(tool.endedAt);
  if(tool.profile)t.profile=tool.profile;
  if(tool.subagentSummary)t.subagentSummary=tool.subagentSummary;
  const card = cardForTool(it, t);
  if (!card) return;
  const elapsed = Math.max(0, Date.now() - Number(card.dataset.startedAt || Date.now()));
  if(isSubagentTool(t)){
    const state=ensureSubagentState(t);
    if(!terminalSubagentPhase(state.phase))state.phase=tool.err?'failed':'completed';
    {const finalPreview=appendSubagentPreview('',tool.output,SUBAGENT_PROGRESS_LIMITS.text);state.finalOutput=finalPreview.value;state.truncated=state.truncated||finalPreview.truncated||!!tool.truncated;}
    state.durationMs=Number(tool.durationMs||state.durationMs||elapsed);
    state.lastActivityAt=Date.now();
    renderSubagentCard(card,t);
    if(subagentAutoCollapse()&&card.dataset.userOverride!=='true'){
      card.dataset.open='false';
      const body=card.querySelector('.card-body');if(body)body.style.display='none';
    }
    updateTurnSummaryForItem(it);
    scrollDown();
    return;
  }
  const ex = tool.execution || null;
  if (ex && tool.name === 'bash') {
    const title = card.querySelector('.card-title .name');
    if (title) title.textContent = shellDisplayName(ex);
  }
  setToolStatus(card, t.err ? 'danger' : 'success', t.err ? __('tool_failed') : __('tool_done'));
  const body = card.querySelector('.card-body');
  let meta = toolResultMeta(t, elapsed);
  if (ex) {
    const bits = [];
    if (typeof ex.exitCode === 'number') bits.push('exit ' + ex.exitCode);
    if (ex.failurePhase) bits.push(ex.failurePhase);
    if (ex.verification && ex.verification !== 'not_verification') bits.push(ex.verification);
    if (bits.length) meta = bits.join(' · ') + ' · ' + meta;
  }
  if (t.err) {
    // Collapsible error detail: three lines preview, click to expand.
    const errBody = el('div', 'err-body');
    if (ex && ex.outputTail) {
      const pre = el('pre', 'err-stderr', ex.outputTail.slice(0, 16000));
      errBody.appendChild(pre);
    }
    const full = String(t.err);
    const preview = full.split('\n').slice(0, 3).join('\n');
    let expanded = false;
    errBody.textContent = preview + (full.length > preview.length ? '\n…' : '');
    errBody.onclick = () => {
      expanded = !expanded;
      errBody.textContent = expanded ? full : preview + (full.length > preview.length ? '\n…' : '');
      errBody.classList.toggle('err-body--expanded', expanded);
    };
    card.appendChild(errBody);
    updateToolMeta(card, meta);
  }
  else if (t.diff || card.dataset.hasDiff) {
    // Diff-driven card: result doesn't replace the diff body; refresh the stat.
    const dv = body.querySelector('.diff-view');
    if (dv) updateToolMeta(card, toolResultMeta(t, elapsed));
  }
  else {
    const output = String(t.output || '');
    if (body) body.textContent = output ? output.slice(0, 2000) + (tool.truncated ? '\n...[truncated]' : '') : __('tool_no_output');
    updateToolMeta(card, toolResultMeta(t, elapsed));
  }
  updateTurnSummaryForItem(it);
  if(t.parentId){const parent=findToolItem(t.parentId);if(parent&&isSubagentTool(parent.tool)){const state=ensureSubagentState(parent.tool);state.lastActivityAt=Date.now();const parentCard=cardForTool(parent.item,parent.tool);if(parentCard)renderSubagentCard(parentCard,parent.tool);}}
  scrollDown();
}
function renderToolProgress(tool) {
  // Reserved sub-agent progress channels (reasonix.subagent.*) are an internal
  // local-frontend contract; they must never land in ordinary tool output.
  if (tool.name && tool.name.indexOf(SUBAGENT_PROGRESS_PREFIX) === 0) {
    const found=findToolItem(tool.id);
    if(!found)return;
    const {item:it,tool:t}=found;
    const state=ensureSubagentState(t);
    state.lastActivityAt=Date.now();
    switch(tool.name){
      case 'reasonix.subagent.status': {
        const phase=String(tool.output||'');
        if(!SUBAGENT_PHASES.has(phase))return;
        state.phase=phase;
        if(terminalSubagentPhase(phase)&&tool.durationMs)state.durationMs=Number(tool.durationMs);
        break;
      }
      case 'reasonix.subagent.reasoning': {const next=appendSubagentPreview(state.reasoning,tool.output,SUBAGENT_PROGRESS_LIMITS.reasoning);state.reasoning=next.value;state.truncated=state.truncated||next.truncated||!!tool.truncated;break;}
      case 'reasonix.subagent.text': {const next=appendSubagentPreview(state.text,tool.output,SUBAGENT_PROGRESS_LIMITS.text);state.text=next.value;state.truncated=state.truncated||next.truncated||!!tool.truncated;break;}
      case 'reasonix.subagent.notice': {const next=appendSubagentPreview(state.notice,tool.output,SUBAGENT_PROGRESS_LIMITS.notice);state.notice=next.value;state.truncated=state.truncated||next.truncated||!!tool.truncated;break;}
      default:return;
    }
    const card=cardForTool(it,t);if(card){renderSubagentCard(card,t);if(terminalSubagentPhase(state.phase)&&subagentAutoCollapse()&&card.dataset.userOverride!=='true'){card.dataset.open='false';const body=card.querySelector('.card-body');if(body)body.style.display='none';}}
    updateTurnSummaryForItem(it);scrollDown();return;
  }
  const found = findTool(tool.id);
  if (!found) return;
  const { item: it, tool: t } = found;
  t.output += tool.output || '';
  const card = cardForTool(it, t);
  if (!card) return;
  const body = card.querySelector('.card-body'); if (!body) return;
  body.style.display=card.dataset.open==='true'?'':'none';
  body.textContent += tool.output || '';
  if(body.textContent.length>4000) body.textContent=body.textContent.slice(-3000);
  updateToolMeta(card,fmtElapsed(Date.now()-Number(card.dataset.startedAt||Date.now()))+' · '+lineCount(body.textContent)+' '+__('tool_lines'));
  updateTurnSummaryForItem(it);
  scrollDown();
}

// Serve renders the structured extension contract as safe native text. It
// never injects extension-provided HTML/Markdown into innerHTML, so a sidecar
// cannot turn a card, form, status, or notification into executable content.
function renderExtensionSurface(p) {
  if(!p)return;
  const lines=[];
  const plugin=p.pluginId||'extension';
  let severity='';
  if(p.status){
    severity=p.status.severity||'';
    lines.push(p.status.label||'');
    if(p.status.detail)lines.push(p.status.detail);
    if(typeof p.status.progress==='number')lines.push(Math.round(p.status.progress*100)+'%');
  } else if(p.card){
    lines.push(p.card.title||'');
    lines.push(p.card.text||p.card.markdown||'');
    (p.card.fields||[]).forEach(f=>lines.push((f.key||'')+': '+(f.value||'')));
    (p.card.actions||[]).forEach(a=>lines.push('['+(a.label||a.actionId||'action')+']'));
  } else if(p.form){
    lines.push(p.form.title||'');
    lines.push(p.form.message||'');
    (p.form.fields||[]).forEach(f=>{
      let detail=(f.label||f.key||'')+(f.required?' *':'');
      if((f.options||[]).length)detail+=' — '+f.options.join(', ');
      lines.push(detail);
    });
  } else if(p.notification){
    severity=p.notification.severity||'';
    lines.push(p.notification.title||'');
    lines.push(p.notification.body||'');
  }
  const text='['+plugin+'] '+lines.filter(Boolean).join('\n');
  const node=el('div','notice'+((severity==='warn'||severity==='error')?' notice--warn':''),text);
  log.appendChild(node);scrollDown();
}

// ── approval ──
function bashCommandPrefix(subject) {
  const command = String(subject || '').trim();
  if (!command || command.includes('`') || command.includes('$(') || /[;|&<>\n]/.test(command)) return '';
  const fields = command.split(/\s+/).filter(Boolean);
  if (fields.length < 2) return '';
  if (dangerousBashCommand(command)) return '';
  const base = fields[0].toLowerCase();
  if (['npm','pnpm','yarn','bun'].includes(base) && fields[1] && fields[1].toLowerCase() === 'run') return fields.length >= 3 ? fields[0]+' '+fields[1]+' '+fields[2]+':*' : '';
  return fields[0]+' '+fields[1]+':*';
}
function dangerousBashCommand(command) {
  return /^rm\s+-[^\s]*[rf][^\s]*\b/.test(command)
    || /^git\s+push\b.*\s--force\b/.test(command)
    || /^git\s+push\b.*\s-f\b/.test(command)
    || /^git\s+reset\s+--hard\b/.test(command)
    || /^git\s+clean\s+-f\b/.test(command)
    || /^chmod\s+(?:-R\s+)?777\b/.test(command)
    || /^chown\b/.test(command)
    || /^sudo\b/.test(command)
    || /^mkfs\b/.test(command)
    || /^dd\s+if=/.test(command)
    || /^fdisk\b/.test(command);
}
// A turn's approval/ask prompts die with the turn: cancel (composer Esc, stop
// button) or turn end must remove the cards and their key listeners, or a
// stale card keeps soliciting a decision for a dead request and the approval
// key listener would resolve it on a later unrelated keypress.
let pendingPrompts=[];
// The approval shelf is a single fixed slot under the transcript (desktop
// footer--decision style): a new request replaces the previous card, whose
// cleanup removes its keydown listener first.
let currentApprovalCleanup = null;
function clearPendingPrompts(){pendingPrompts.splice(0).forEach(fn=>fn());}
function focusDecisionPrompt(){
  const prompt=document.querySelector('.approval, .ask');
  if(prompt){const target=prompt.matches('.approval')?prompt:prompt.querySelector('button:not(:disabled),input:not(:disabled),textarea:not(:disabled)')||prompt;target.focus?.({preventScroll:true});prompt.scrollIntoView?.({block:'nearest'});}
}
function ordinaryOverlayAllowed(){
  if(!decisionInteractionLocked)return true;
  showNotice(__('workspace_decision_pending'),'warn');focusDecisionPrompt();return false;
}
function closeOrdinaryModals(){
  closeImageViewer();document.getElementById('rewind-overlay')?.remove();
  ['stats-modal','branches-modal','models-modal','delete-modal'].forEach(id=>{const node=document.getElementById(id);if(node)node.style.display='none';});pendingDeleteSession=null;
}
function prepareOrdinaryOverlay(){
  if(!ordinaryOverlayAllowed())return false;
  closeWorkspacePanel({restoreFocus:false});closeSettings({restoreFocus:false});closeOrdinaryModals();return true;
}
function closeOrdinaryOverlaysForDecision(){
  closeWorkspacePanel({restoreFocus:false});
  closeSettings({preserveDraft:true,restoreFocus:false});
  closeOrdinaryModals();
  closeSlashMenu();
  const modelMenu=document.getElementById('modelsw-menu'),effort=document.getElementById('effort-menu'),taskMenu=document.getElementById('task-mode-menu');
  if(modelMenu)modelMenu.style.display='none';if(effort)effort.style.display='none';if(taskMenu)taskMenu.style.display='none';
}
function beginDecisionInteraction(){
  decisionInteractionLocked=true;document.body.classList.add('decision-pending');closeOrdinaryOverlaysForDecision();
}
function endDecisionInteraction(){
  if(waitingPrompt!==null)return;
  decisionInteractionLocked=false;document.body.classList.remove('decision-pending');
}
function showApproval(a) {
  // Single-slot shelf: retire any pending approval card (its cleanup removes
  // the keydown listener and unregisters from pendingPrompts) before showing
  // the new request.
  if (currentApprovalCleanup) currentApprovalCleanup();
  currentApprovalCleanup = null;
  pendingApprovalLabel = a.tool || '';
  waitingPrompt = 'approval';
  beginDecisionInteraction();
  waitPause();
  updateRunStrip();
  const d = el('div', 'approval');
  d.tabIndex = -1;
  // Plan approvals (exit_plan_mode) are fresh human decisions: a dedicated
  // card like the desktop ApprovalModal's isPlanApproval branch — approve the
  // plan execution or deny it. No session/persist grants apply.
  const isPlan = a.tool === 'exit_plan_mode';
  const prefix = a.tool === 'bash' ? bashCommandPrefix(a.subject) : '';
  const hasPrefix = prefix !== '';
  const prefixRule = hasPrefix ? 'Bash(' + prefix + ')' : '';
  // Actions mirror desktop's prompt-shelf tool approval: fixed labels + a
  // small desc line, never the command text (the subject code block above
  // shows the full command). The rule (e.g. Bash(prefix)) is kept on the
  // button's title tooltip. Payloads keep the session/persist semantics the
  // controller expects.
  const actions = isPlan
    ? [
        { label: __('approval_plan_approve'), desc: __('approval_plan_approve_desc'), payload: { allow: true, session: false }, danger: false, rule: '' },
        { label: __('deny'), desc: __('deny_desc'), payload: { allow: false, session: false }, danger: true, rule: '' },
      ]
    : [
        { label: __('allow_once'), desc: __('allow_once_desc'), payload: { allow: true, session: false }, danger: false, rule: '' },
      ];
  if (!isPlan) {
    if (hasPrefix) {
      actions.push({ label: __('session'), desc: __('session_desc'), payload: { allow: true, session: true, scope: 'prefix' }, danger: false, rule: prefixRule });
      actions.push({ label: __('persist_tool'), desc: __('persist_desc'), payload: { allow: true, session: true, persist: true, scope: 'prefix' }, danger: false, rule: prefixRule });
    } else {
      actions.push({ label: __('session'), desc: __('session_desc'), payload: { allow: true, session: true }, danger: false, rule: '' });
      actions.push({ label: __('persist_tool'), desc: __('persist_desc'), payload: { allow: true, session: true, persist: true }, danger: false, rule: '' });
    }
    actions.push({ label: __('deny'), desc: __('deny_desc'), payload: { allow: false, session: false }, danger: true, rule: '' });
  }

  const card = el('div', 'approval__card');
  // Head: amber tool badge + title + optional Details toggle for long reasons.
  const head = el('div', 'approval__head');
  head.appendChild(el('span', 'approval__badge', isPlan ? __('approval_plan_badge') : a.tool));
  head.appendChild(el('span', 'approval__title', isPlan ? __('approval_plan_title') : __('approval_title')));
  const reason = a.reason ? String(a.reason) : '';
  const reasonLong = reason.length > 160;
  let reasonEl = null;
  if (reason) {
    const detailsBtn = el('button', 'approval__head-btn', reasonLong ? __('approval_details') : __('approval_hide'));
    detailsBtn.type = 'button';
    head.appendChild(detailsBtn);
    reasonEl = el('div', 'approval__reason', reason);
    if (reasonLong) reasonEl.style.display = 'none';
    detailsBtn.onclick = () => {
      const open = reasonEl.style.display !== 'none';
      reasonEl.style.display = open ? 'none' : '';
      detailsBtn.textContent = open ? __('approval_details') : __('approval_hide');
    };
  }
  card.appendChild(head);
  // Scrollable body: meta + subject + reason share one capped, scrollable
  // region (desktop prompt-shelf--decision body), so a huge subject or reason
  // never pushes the action list / confirm bar out of view.
  const scroll = el('div', 'approval__scroll');
  const meta = a.subject ? compactText(a.subject, 90) : (reason ? compactText(reason, 90) : '');
  if (meta) scroll.appendChild(el('div', 'approval__meta', meta));
  if (a.subject) scroll.appendChild(el('div', 'approval__subject', a.subject));
  if (reasonEl) scroll.appendChild(reasonEl);
  card.appendChild(scroll);

  // Actions list (single column, select-then-confirm).
  let selected = 0;
  const list = el('div', 'approval__actions');
  const actionEls = actions.map((act, i) => {
    const btn = el('button', 'approval-action' + (act.danger ? ' approval-action--danger' : ''));
    btn.type = 'button';
    if (act.rule) btn.title = act.rule; // e.g. Bash(npm run build:*)
    btn.innerHTML = '<span class="approval-action__key">' + (i + 1) + '</span>'
      + '<span class="approval-action__text"><span class="approval-action__label">' + escHtml(act.label) + '</span>'
      + '<span class="approval-action__desc">' + escHtml(act.desc) + '</span></span>';
    btn.onclick = () => { selected = i; renderSelect(); };
    list.appendChild(btn);
    return btn;
  });
  card.appendChild(list);

  // Bottom confirm bar.
  const confirm = el('div', 'approval__confirm');
  confirm.appendChild(el('span', 'approval__hint', __('approval_hint')));
  const confirmBtn = el('button', 'approval__confirm-btn', '');
  confirmBtn.type = 'button';
  confirm.appendChild(confirmBtn);
  card.appendChild(confirm);
  d.appendChild(card);
  approvalSlot.appendChild(d);

  function renderSelect() {
    actionEls.forEach((b, i) => b.classList.toggle('approval-action--selected', i === selected));
    confirmBtn.textContent = actions[selected].label;
    confirmBtn.classList.toggle('approval__confirm-btn--danger', actions[selected].danger);
  }
  renderSelect();

  // The advertised number keys need a non-text-entry target (isPlainKey), but
  // the composer is autofocused, so in the default state every shortcut would
  // just type into the draft. Move focus onto the card so the keys route here —
  // unless the user is mid-entry (a non-empty composer draft, or any other
  // text field such as the session search): stealing focus there would turn
  // their next keystroke into an unintended approval. Cleanup hands focus back
  // to the composer only when it still sits inside the card, so a user who
  // deliberately moved elsewhere is not yanked back.
  if (!isTextEntry(document.activeElement) || (document.activeElement === input && !input.value)) d.focus({ preventScroll: true });
  const cleanup = () => { pendingPrompts = pendingPrompts.filter(f => f !== cleanup); if (waitingPrompt === 'approval') { waitingPrompt = null; waitResume(); updateRunStrip(); } const hadFocus = d.contains(document.activeElement); d.remove(); document.removeEventListener('keydown', onkey); if (currentApprovalCleanup === cleanup) currentApprovalCleanup = null; endDecisionInteraction(); if (hadFocus) input.focus(); };
  currentApprovalCleanup = cleanup;
  pendingPrompts.push(cleanup);
  const resolve = payload => { post('/approve', Object.assign({ id: a.id }, payload)); cleanup(); };
  const commit = () => resolve(actions[selected].payload);
  confirmBtn.onclick = commit;
  // Number keys select an action, Enter commits, Esc denies, arrows cycle.
  // A consumed shortcut must preventDefault: cleanup refocuses the composer
  // synchronously, and the keystroke's default text insertion runs after
  // keydown — without it the approving key lands in the draft as a stray char.
  const onkey = e => {
    if (!isPlainKey(e)) return;
    const n = Number(e.key);
    if (Number.isInteger(n) && n >= 1 && n <= actions.length) { e.preventDefault(); selected = n - 1; renderSelect(); return; }
    if (e.key === 'ArrowDown') { e.preventDefault(); selected = Math.min(selected + 1, actions.length - 1); renderSelect(); return; }
    if (e.key === 'ArrowUp') { e.preventDefault(); selected = Math.max(selected - 1, 0); renderSelect(); return; }
    if (e.key === 'Enter') { e.preventDefault(); commit(); return; }
    if (e.key === 'Escape') { e.preventDefault(); resolve({ allow: false, session: false }); return; }
  };
  document.addEventListener('keydown', onkey);
}

// ── ask ──
// ── ask decision shelf (desktop .prompt-shelf parity) ──
let askQueue = [], askActive = false, askAnswers = [], askCur = 0, askFocused = 0;
const askSlot = $('#ask-slot');
function showAsk(ask) {
  askQueue.push(ask);
  if (!askActive) {
    askActive = true;
    askAnswers = []; askCur = 0; askFocused = 0;
    waitingPrompt = 'ask';
    beginDecisionInteraction();
    waitPause();
    updateRunStrip();
  }
  if (askQueue.length === 1) renderAskCard();
}
function renderAskCard() {
  const ask = askQueue[0];
  if (!askSlot || !ask) return;
  if (askSlot._askCleanup) { askSlot._askCleanup(); askSlot._askCleanup = null; }
  askSlot.innerHTML = '';
  const d = el('div', 'ask');
  const q = ask.questions[askCur];
  const isLast = askCur >= ask.questions.length - 1;
  const hasMulti = ask.questions.length > 1;
  const n = q.options.length;
  const customRow = n, skipRow = n + 1, rowCount = n + 2;
  let askRow = 0, customOpen = false, customText = '';
  const selected = new Set();
  // head: title + badges + stop-task action (desktop AskCard parity)
  const head = el('div', 'ask__head');
  head.appendChild(el('span', 'ask__title', __('ask_title')));
  if (q.header) head.appendChild(el('span', 'ask__badge', q.header));
  if (hasMulti) head.appendChild(el('span', 'ask__progress', __('ask_progress').replace('{n}', String(askCur + 1)).replace('{m}', String(ask.questions.length))));
  const stop = el('button', 'ask__stop', __('ask_stop_task'));
  stop.type = 'button';
  stop.onclick = stopAsk;
  head.appendChild(stop);
  d.appendChild(head);
  d.appendChild(el('div', 'ask__prompt', q.prompt));
  // rows: options + "other answer" + "skip and keep chatting"
  const rows = el('div', 'ask__rows');
  const optEls = [];
  const prev = askAnswers[askCur];
  q.options.forEach((o, i) => {
    const opt = el('button', 'ask__opt');
    opt.type = 'button';
    opt.innerHTML = '<span class="ask__num">' + (n <= 9 ? (i + 1) : '') + '</span><div><div class="ask__opt-label">' + escHtml(o.label) + '</div>' + (o.description ? '<div class="ask__opt-desc">' + escHtml(o.description) + '</div>' : '') + '</div>';
    opt.onclick = () => selectRow(i);
    rows.appendChild(opt);
    optEls.push(opt);
  });
  const customOpt = el('button', 'ask__opt ask__opt--custom');
  customOpt.type = 'button';
  customOpt.innerHTML = '<span class="ask__num"></span><div><div class="ask__opt-label">' + escHtml(__('ask_custom_answer')) + '</div><div class="ask__opt-desc">' + escHtml(__('ask_custom_answer_desc')) + '</div></div>';
  customOpt.onclick = () => selectRow(customRow);
  rows.appendChild(customOpt);
  optEls.push(customOpt);
  const skipOpt = el('button', 'ask__opt ask__opt--skip');
  skipOpt.type = 'button';
  skipOpt.innerHTML = '<span class="ask__num"></span><div><div class="ask__opt-label">' + escHtml(__('ask_just_chat')) + '</div><div class="ask__opt-desc">' + escHtml(__('ask_just_chat_desc')) + '</div></div>';
  skipOpt.onclick = () => selectRow(skipRow);
  rows.appendChild(skipOpt);
  optEls.push(skipOpt);
  d.appendChild(rows);
  // custom answer input (expanded on demand)
  const customBox = el('div', 'ask__custom-row');
  customBox.style.display = 'none';
  const cin = el('input');
  cin.type = 'text';
  cin.placeholder = __('ask_custom_placeholder');
  cin.addEventListener('keydown', e => {
    if (e.key === 'Enter' && cin.value.trim()) { e.preventDefault(); confirmSelected(); }
    e.stopPropagation();
  });
  cin.addEventListener('input', () => { customText = cin.value; updateConfirm(); });
  customBox.appendChild(cin);
  d.appendChild(customBox);
  // crumbs: answered-summary capsules of prior questions
  const crumbs = el('div', 'ask__crumbs');
  d.appendChild(crumbs);
  // confirm bar: back + hint + dynamic confirm on one row, same look
  const bar = el('div', 'ask__confirmbar');
  const hint = el('span', 'ask__hint', __('ask_select_hint'));
  const actions = el('div', 'ask__actions');
  if (hasMulti && askCur > 0) {
    const back = el('button', 'ask__back', __('ask_back'));
    back.type = 'button';
    back.onclick = () => goBack();
    actions.appendChild(back);
  }
  const confirm = el('button', 'ask__confirm', __('ask_next'));
  confirm.type = 'button';
  confirm.disabled = true;
  confirm.onclick = confirmSelected;
  actions.appendChild(confirm);
  bar.appendChild(hint);
  bar.appendChild(actions);
  d.appendChild(bar);
  askSlot.appendChild(d);
  // restore previously given answers (Back navigation)
  if (prev && prev.selected.length) {
    prev.selected.forEach(s => { const idx = q.options.findIndex(o => o.label === s); if (idx >= 0) selected.add(idx); });
    if (!q.multi && selected.size) askRow = [...selected][0];
  }
  function selectRow(i) {
    if (i < n) {
      if (q.multi) { selected.has(i) ? selected.delete(i) : selected.add(i); }
      else { selected.clear(); selected.add(i); }
      customOpen = false; customText = '';
    } else if (i === customRow) {
      customOpen = true;
      if (!q.multi) selected.clear();
    } else if (i === skipRow) {
      customOpen = false; customText = '';
      if (!q.multi) selected.clear();
    }
    askRow = i;
    updateRows();
  }
  function updateRows() {
    optEls.forEach((oe, j) => {
      const on = j < n ? selected.has(j) : (j === customRow ? customOpen : askRow === skipRow);
      oe.classList.toggle('ask__opt--selected', on);
      oe.classList.toggle('ask__opt--focused', j === askRow);
    });
    customBox.style.display = customOpen ? '' : 'none';
    if (customOpen) cin.focus();
    updateConfirm();
  }
  function updateConfirm() {
    confirm.disabled = !canConfirm();
    confirm.textContent = askRow === skipRow ? __('ask_just_chat') : isLast ? __('submit') : __('ask_next');
  }
  function canConfirm() {
    if (askRow === skipRow) return true;
    if (askRow === customRow) return customText.trim() !== '';
    if (q.multi) return selected.size > 0;
    return askRow < n || selected.size > 0;
  }
  function confirmSelected() {
    if (!canConfirm()) return;
    if (askRow === skipRow) {
      // Skip and keep chatting: empty answers for the whole ask.
      finishAsk(ask.questions.map(qq => ({ questionId: qq.id, selected: [] })));
      return;
    }
    if (askRow === customRow) { answerCurrent([customText.trim()]); return; }
    if (q.multi) { answerCurrent(Array.from(selected).map(i => q.options[i].label)); return; }
    const label = askRow < n ? q.options[askRow].label : (selected.size ? q.options[[...selected][0]].label : '');
    if (label) answerCurrent([label]);
  }
  function goBack() { if (askCur > 0) { askCur--; renderAskCard(); } }
  // crumbs content: labels of all answered questions before the current one.
  crumbs.innerHTML = '';
  for (let i = 0; i < askCur; i++) {
    const a = askAnswers[i];
    if (a && a.selected.length) crumbs.appendChild(el('span', 'ask__crumb', (i + 1) + '. ' + escHtml(a.selected.join(', '))));
  }
  const onkey = e => {
    if (document.activeElement === cin) return;
    if (e.key === 'Escape') { e.preventDefault(); stopAsk(); }
    else if (e.key === 'Enter') { e.preventDefault(); confirmSelected(); }
    else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
      e.preventDefault();
      const dir = e.key === 'ArrowDown' ? 1 : -1;
      askRow = (askRow + dir + rowCount) % rowCount;
      updateRows();
    } else if (e.key === 'ArrowLeft' || e.key === 'Backspace') {
      if (askCur > 0) { e.preventDefault(); goBack(); }
    } else if (/^[1-9]$/.test(e.key)) {
      const i = +e.key - 1;
      if (i < n) { e.preventDefault(); selectRow(i); }
    }
  };
  document.addEventListener('keydown', onkey);
  askSlot._askCleanup = () => { document.removeEventListener('keydown', onkey); askSlot._askCleanup = null; };
  if (document.activeElement === input && optEls[0]) optEls[0].focus();
}
// stopAsk cancels the running task and dismisses every queued question
// (desktop parity: Esc = stop task, not skip).
function stopAsk() {
  if (askSlot) { if (askSlot._askCleanup) askSlot._askCleanup(); askSlot.innerHTML = ''; }
  askQueue = [];
  askActive = false;
  waitingPrompt = null;
  waitResume();
  updateRunStrip();
  endDecisionInteraction();
  post('/cancel');
  input.focus();
}
function answerCurrent(selected) {
  const ask = askQueue[0];
  if (!ask) return;
  const q = ask.questions[askCur];
  askAnswers[askCur] = { questionId: q.id, selected: Array.from(selected) };
  if (askSlot && askSlot._askCleanup) askSlot._askCleanup();
  if (askCur < ask.questions.length - 1) { askCur++; renderAskCard(); }
  else finishAsk(askAnswers);
}
function finishAsk(answers) {
  const ask = askQueue[0];
  if (!ask) return;
  if (askSlot) askSlot.innerHTML = '';
  askQueue.shift();
  post('/answer', { id: ask.id, answers });
  if (askQueue.length > 0) { renderAskCard(); }
  else {
    askActive = false;
    waitingPrompt = null;
    waitResume();
    updateRunStrip();
    endDecisionInteraction();
    input.focus();
  }
}

// ── compaction ──
function renderCompactionItem(c) {
  const d=el('div','compaction');
  if(c.summary){const head=el('div','compaction__head');head.innerHTML='<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="4 14 10 14 10 20"/><polyline points="20 10 14 10 14 4"/><line x1="14" y1="10" x2="21" y2="3"/><line x1="3" y1="21" x2="10" y2="14"/></svg>';head.appendChild(el('span','compaction__title',__('compacted')));head.appendChild(el('span','',c.messages+' '+__('messages')));const body=el('div','compaction__body',c.summary);head.onclick=()=>body.classList.toggle('compaction__body--open');d.appendChild(head);d.appendChild(body);}
  else d.textContent=__('compacting');
  return d;
}
function showCompaction(c) {
  const it = { id: genItemId(), kind: 'compaction', summary: c.summary || '', messages: c.messages || 0 };
  items.push(it);
  const d = renderCompactionItem(it); if (d) { log.appendChild(d); scrollDown(); }
}

// ── rewind picker ──
let rewindCheckpoints = [];
let rewindStage = 0, rewindSelected = 0, rewindScope = 0;
const SCOPES = [
  {key:'b',label:__('scope_both'),scope:'both'},
  {key:'c',label:__('scope_conversation'),scope:'conversation'},
  {key:'d',label:__('scope_code'),scope:'code'},
  {key:'f',label:__('scope_fork'),scope:'fork'},
  {key:'s',label:__('scope_sumfrom'),scope:'sumfrom'},
  {key:'u',label:__('scope_sumupto'),scope:'sumupto'},
];

function openRewindPicker() {
  if(!prepareOrdinaryOverlay())return;
  fetch('/checkpoints').then(r=>r.json()).then(cps=>{
    checkpointCount=Array.isArray(cps)?cps.length:0; updateActionAvailability();
    if(!cps||cps.length===0){showNotice(__('no_checkpoints'),'warn');return;}
    rewindCheckpoints=cps; rewindStage=0; rewindSelected=0; rewindScope=0;
    renderRewindPicker();
  }).catch(()=>{});
}
function refreshCheckpointAvailability(){
  fetch('/checkpoints').then(r=>r.json()).then(cps=>{
    checkpointCount=Array.isArray(cps)?cps.length:0;
    updateActionAvailability();
  }).catch(()=>{});
}

function renderRewindPicker() {
  let overlay=$('#rewind-overlay'); if(overlay)overlay.remove();
  overlay=el('div','rewind-overlay'); overlay.id='rewind-overlay';
  const picker=el('div','rewind-picker');

  if(rewindStage===0){
    picker.innerHTML='<div class="rewind-picker__head"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg> '+__('rewind_title')+'</div>';
    const list=el('div','rewind-picker__list');
    rewindCheckpoints.forEach((cp,i)=>{
      const item=el('div','rewind-picker__item'+(i===rewindSelected?' rewind-picker__item--active':''));
      item.innerHTML='<span class="rewind-picker__turn">#'+cp.turn+'</span><span class="rewind-picker__prompt">'+escHtml((cp.prompt||'').slice(0,80))+'</span><span class="rewind-picker__files">'+cp.files+' '+__('files')+'</span>';
      item.onclick=()=>{rewindSelected=i;renderRewindPicker();};
      list.appendChild(item);
    });
    picker.appendChild(list);
    picker.appendChild(el('div','rewind-picker__foot','<span>'+__('nav_jk')+'</span><span>'+__('nav_enter_esc')+'</span>'));
  } else {
    const cp=rewindCheckpoints[rewindSelected];
    picker.innerHTML='<div class="rewind-picker__head"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg> '+__('action_title').replace('#{turn}',cp.turn)+'</div>';
    const scopes=el('div','rewind-picker__scopes');
    SCOPES.forEach((s,i)=>{
      const item=el('div','rewind-picker__scope'+(i===rewindScope?' rewind-picker__scope--active':''));
      item.innerHTML='<span class="rewind-picker__scope-key">'+s.key+'</span>'+s.label;
      item.onclick=()=>{rewindScope=i;applyRewind();};
      scopes.appendChild(item);
    });
    picker.appendChild(scopes);
    picker.appendChild(el('div','rewind-picker__foot','<span>'+__('nav_keys')+'</span><span>'+__('nav_apply_esc')+'</span>'));
  }

  overlay.appendChild(picker);
  overlay.onclick=e=>{if(e.target===overlay)overlay.remove();};
  document.body.appendChild(overlay);
}

function applyRewind() {
  const cp=rewindCheckpoints[rewindSelected]; const sc=SCOPES[rewindScope];
  document.getElementById('rewind-overlay')?.remove();
  if(sc.scope==='fork'){post('/fork',{turn:cp.turn,name:''});}
  else if(sc.scope==='sumfrom'){post('/summarize',{turn:cp.turn,mode:'from'});}
  else if(sc.scope==='sumupto'){post('/summarize',{turn:cp.turn,mode:'upto'});}
  else{post('/rewind',{turn:cp.turn,scope:sc.scope});}
}

document.addEventListener('keydown',e=>{
  const overlay=$('#rewind-overlay'); if(!overlay)return;
  // The picker is modal and runs in the capture phase: plain keys it owns are
  // swallowed before they reach the composer underneath — Enter would
  // otherwise both send the composer draft and advance the picker, and
  // letters would type into the hidden draft. Modifier chords
  // (copy/reload/devtools), IME composition, and non-printable keys (F5, Tab)
  // pass through — a bare Cmd+C must not hit a scope key and apply a rewind.
  if(e.ctrlKey||e.metaKey||e.altKey||e.isComposing||e.key==='Process')return;
  if(!(e.key==='Escape'||e.key==='Enter'||e.key==='ArrowDown'||e.key==='ArrowUp'||e.key.length===1))return;
  e.preventDefault();e.stopPropagation();
  if(e.key==='Escape'){if(rewindStage===0)overlay.remove();else{rewindStage=0;renderRewindPicker();}return;}
  if(rewindStage===0){
    if(e.key==='j'||e.key==='ArrowDown'){rewindSelected=Math.min(rewindSelected+1,rewindCheckpoints.length-1);renderRewindPicker();}
    if(e.key==='k'||e.key==='ArrowUp'){rewindSelected=Math.max(rewindSelected-1,0);renderRewindPicker();}
    if(e.key==='Enter'){rewindStage=1;rewindScope=0;renderRewindPicker();}
  } else {
    if(e.key==='j'||e.key==='ArrowDown'){rewindScope=Math.min(rewindScope+1,SCOPES.length-1);renderRewindPicker();}
    if(e.key==='k'||e.key==='ArrowUp'){rewindScope=Math.max(rewindScope-1,0);renderRewindPicker();}
    if(e.key==='Enter'){applyRewind();}
    const idx=SCOPES.findIndex(s=>s.key===e.key); if(idx>=0){rewindScope=idx;applyRewind();}
  }
},true);

// ── slash menu ──
let slashOpen=false, slashIndex=0, slashFiltered=[], slashArgMode=false, slashArgItems=[], slashArgCmd='';
function updateSlashMenu() {
  const v=input.value;
  if(!v.startsWith('/')){closeSlashMenu();return;}
  const space=v.indexOf(' ');
  if(space>=0){
    // Argument completion (desktop ArgMenu): /model <ref>, /effort <level>
    const cmd=v.slice(0,space);
    if(cmd==='/model'){
      const q=v.slice(space+1).trim().toLowerCase();
      ensureModelsLoaded().then(()=>{
        if(!input.value.startsWith('/model '))return; // user moved on while loading
        slashArgItems=uniqueModelRefs().filter(r=>r.toLowerCase().includes(q));
        if(slashArgItems.length===0){closeSlashMenu();return;}
        slashArgCmd='/model';
        slashOpen=true; slashIndex=0; slashArgMode=true;
        renderSlashArgs(q);
      });
      return;
    }
    if(cmd==='/effort'||cmd==='/thinking'){
      const q=v.slice(space+1).trim().toLowerCase();
      ensureEffortLoaded().then(()=>{
        if(!input.value.startsWith(cmd+' '))return; // user moved on while loading
        if(!effortState||!effortState.supported||!Array.isArray(effortState.levels)){closeSlashMenu();return;}
        slashArgItems=effortState.levels.filter(l=>l.toLowerCase().includes(q));
        if(slashArgItems.length===0){closeSlashMenu();return;}
        slashArgCmd=cmd;
        slashOpen=true; slashIndex=0; slashArgMode=true;
        renderSlashArgs(q);
      });
      return;
    }
    closeSlashMenu();return;
  }
  slashArgMode=false;
  const q=v.slice(1).toLowerCase();
  slashFiltered=SLASH_CMDS.filter(c=>{
    const hay=[c.cmd,c.sig,c.desc,slashGroupLabel(c.group)].join(' ').toLowerCase();
    return hay.includes(q);
  }).sort((a,b)=>{
    const ap=a.cmd.startsWith(q)?0:1, bp=b.cmd.startsWith(q)?0:1;
    return ap-bp || SLASH_CMDS.indexOf(a)-SLASH_CMDS.indexOf(b);
  });
  if(slashFiltered.length===0){closeSlashMenu();return;}
  slashOpen=true; slashIndex=0;
  renderSlashMenu();
}
function renderSlashMenu() {
  let menu=$('#slash-menu'); if(!menu){menu=el('div','slash-menu');menu.id='slash-menu';slashAnchor.appendChild(menu);}
  const query=input.value.slice(1);
  menu.innerHTML='<div class="slash-menu__head"><span>'+escHtml(__('command_palette'))+'</span><span class="slash-menu__query">/'+escHtml(query)+'</span></div><div class="slash-menu__list" id="slash-menu-list" role="listbox"></div><div class="slash-menu__foot"><span>'+escHtml(__('command_nav'))+'</span><span>'+slashFiltered.length+'</span></div>';
  const list=$('#slash-menu-list');
  let lastGroup='';
  slashFiltered.forEach((c,i)=>{
    if(c.group!==lastGroup){
      lastGroup=c.group;
      const group=el('div','slash-menu__group',slashGroupLabel(c.group));
      list.appendChild(group);
    }
    const item=el('button','slash-menu__item'+(i===slashIndex?' slash-menu__item--active':''));
    item.type='button';
    item.setAttribute('role','option');
    item.setAttribute('aria-selected',i===slashIndex?'true':'false');
    item.innerHTML='<span class="slash-menu__name">/'+escHtml(c.cmd)+'</span><span class="slash-menu__desc">'+escHtml(c.desc)+'</span><span class="slash-menu__pill'+(c.danger?' slash-menu__pill--danger':'')+'">'+escHtml(c.danger?__('danger'):slashGroupLabel(c.group))+'</span><span class="slash-menu__sig">'+escHtml(c.sig)+'</span>';
    item.onmouseenter=()=>{if(slashIndex!==i){slashIndex=i;renderSlashMenu();}};
    item.onclick=()=>{input.value='/'+c.cmd+' ';closeSlashMenu();input.focus();};
    list.appendChild(item);
  });
  const active=menu.querySelector('.slash-menu__item--active');
  if(active)active.scrollIntoView({block:'nearest'});
}
function renderSlashArgs(q) {
  let menu=$('#slash-menu'); if(!menu){menu=el('div','slash-menu');menu.id='slash-menu';slashAnchor.appendChild(menu);}
  const cmdName=slashArgCmd||'/model';
  const isModel=cmdName==='/model';
  menu.innerHTML='<div class="slash-menu__head"><span>'+escHtml(__('command_palette'))+'</span><span class="slash-menu__query">'+escHtml(cmdName)+' '+escHtml(q)+'</span></div><div class="slash-menu__list" id="slash-menu-list" role="listbox"></div><div class="slash-menu__foot"><span>'+escHtml(__('command_nav'))+'</span><span>'+slashArgItems.length+'</span></div>';
  const list=$('#slash-menu-list');
  slashArgItems.forEach((ref,i)=>{
    const item=el('button','slash-menu__item'+(i===slashIndex?' slash-menu__item--active':''));
    item.type='button';
    item.setAttribute('role','option');
    item.setAttribute('aria-selected',i===slashIndex?'true':'false');
    const isCurrent=isModel?ref===currentModelRefLabel():ref===(effortState.current||'auto');
    item.innerHTML='<span class="slash-menu__name" style="--slash-name:var(--fg)">'+escHtml(ref)+'</span><span class="slash-menu__desc">'+escHtml(isModel?(ref.split('/')[0]||''):__('effort_title'))+'</span><span class="slash-menu__pill'+(isCurrent?' slash-menu__pill--current':'')+'">'+escHtml(isCurrent?__('current'):'')+'</span><span class="slash-menu__sig"></span>';
    item.onmouseenter=()=>{if(slashIndex!==i){slashIndex=i;renderSlashArgs(q);}};
    item.onclick=()=>{input.value=cmdName+' '+ref;closeSlashMenu();input.focus();};
    list.appendChild(item);
  });
  const active=menu.querySelector('.slash-menu__item--active');
  if(active)active.scrollIntoView({block:'nearest'});
}
function closeSlashMenu(){slashOpen=false;const m=$('#slash-menu');if(m)m.remove();}
function acceptSlash(){if(!slashOpen)return;
  if(slashArgMode){
    const it=slashArgItems[slashIndex];
    if(it){input.value=(slashArgCmd||'/model')+' '+it;}
    closeSlashMenu();input.focus();return;
  }
  const c=slashFiltered[slashIndex];if(c){input.value='/'+c.cmd+' ';}closeSlashMenu();input.focus();}

// ── slash argument completion helpers ──
function ensureModelsLoaded() {
  if(modelsCache.length>0)return Promise.resolve();
  return fetch('/models').then(r=>r.json()).then(d=>{modelsCache=Array.isArray(d?.models)?d.models:[];}).catch(()=>{modelsCache=[];});
}
// Effort levels are small and can change with the active model, so the /effort
// menu always fetches fresh state (also keeps effortState/button in sync). One
// in-flight promise is shared across keystrokes; a failure keeps the last good
// effortState instead of clobbering it (the effort button stays usable).
let effortFetch=null;
function ensureEffortLoaded() {
  if(!effortFetch){
    effortFetch=fetch('/effort').then(r=>r.json()).then(d=>{effortState=d||{};}).catch(()=>{}).finally(()=>{effortFetch=null;});
  }
  return effortFetch;
}
// Unique model refs matching /model text output semantics: same-named models
// across provider aliases collapse to one ref (active/default entry wins).
function uniqueModelRefs() {
  const byModel=new Map();
  for(const m of modelsCache){
    const key=m.model||m.ref||'';
    if(!key)continue;
    const ex=byModel.get(key);
    if(!ex)byModel.set(key,m);
    else if((m.active||m.default)&&!(ex.active||ex.default))byModel.set(key,m);
  }
  return [...byModel.values()].map(m=>m.ref||m.model||'');
}
function currentModelRefLabel(){const c=modelsCache.find(m=>m.active);return c?(c.ref||c.model||''):'';}

// ── SSE ──
let es;
function connectEvents(){
es=new EventSource('/events');
es.onopen=()=>{setConnState('connected');fetchStatus();fetchTodos();};
es.onmessage=ev=>{setConnState('connected');
  if(historyPending)return; // history rebuild owns the transcript; skip gap events
  const e=JSON.parse(ev.data);
  if(e.kind!=='retrying')clearRetrying();
  switch(e.kind){
    case 'turn_started': setRunning(true); clearPendingPrompts(); finalizeMsg(); currentTurn++; turnArgChars = 0; todosDismissed=false; break;
    case 'reasoning': if(e.reasoning||e.text){turnOutputChars+=(e.reasoning||e.text).length;beginModelActivity();} appendReasoning(e.reasoning||e.text||''); break;
    case 'text': if(e.text){turnOutputChars+=e.text.length;beginModelActivity();} appendText(e.text||''); break;
    case 'message': finalizeMsg(); break;
    case 'tool_dispatch': if(e.tool){renderToolDispatch(e.tool); if(!e.tool.parentId){if(e.tool.partial){beginModelActivity();}else{endModelActivity();}} if (e.tool.argChars && e.tool.argChars > 0) turnArgChars = e.tool.argChars;} break;
    case 'tool_result': if(e.tool){renderToolResult(e.tool);if(e.tool.name==='todo_write'&&!e.tool.parentId&&!e.tool.err){try{const ts=parseTodos(e.tool.args);if(ts.length){todosState=ts;renderTodoPanel();}}catch{}}} break;
    case 'tool_progress': if(e.tool)renderToolProgress(e.tool); break;
    case 'usage':
      // Usage never enters the transcript (matches desktop): accumulate into
      // the session stats that drive the sidebar metrics and the stats modal.
      if(e.usage){turnTokens+=e.usage.completionTokens||0;
        // Desktop parity: only executor usage closes the model window and
        // feeds the TPS output total; subagent/planner usage still counts
        // toward the turn total (turnTokens accumulates unconditionally).
        const execUsage=!e.usage.source||e.usage.source==='executor';
        if(execUsage){turnOutputTokens+=e.usage.completionTokens||0;
          turnOutputCharsAtUsage = turnOutputChars; // desktop parity: usage snapshot closes the char estimate
          endModelActivity(); // usage closes the model-active window (desktop parity)
          turnArgChars = 0; // desktop parity: usage closes the streaming-args estimate
        }
        cumulativeTokens+=e.usage.totalTokens||0;
        cumulativeCacheHit+=e.usage.cacheHitTokens||0;
        cumulativeCacheMiss+=e.usage.cacheMissTokens||0;
        updateRunStrip();
      } break;
    case 'notice': { if(attachAuditNotice(e)){scrollDown();break;} const d = el('div','notice'+(e.level==='warn'?' notice--warn':''),(e.level==='warn'?'! ':'')+(e.text||'')); const it = { id: genItemId(), kind: 'notice', text: e.text || '', level: e.level, turn: currentTurn }; items.push(it); appendItem(it, d); scrollDown(); break; }
    case 'phase': { finalizeMsg(); const d = el('div','phase',e.text||''); const it = { id: genItemId(), kind: 'phase', text: e.text || '', turn: currentTurn }; items.push(it); appendItem(it, d); scrollDown(); break; }
    case 'approval_request': if(e.approval)showApproval(e.approval); break;
    case 'ask_request': if(e.ask)showAsk(e.ask); break;
    case 'extension_surface': if(e.extension)renderExtensionSurface(e.extension); break;
    case 'extension_status': if(e.extension)renderExtensionSurface(e.extension); break;
    case 'compaction_started': showCompaction({trigger:e.compaction?.trigger}); break;
    case 'compaction_done': showCompaction(e.compaction||{}); fetchStatus(); break;
    case 'retrying': setRetrying(e.retryAttempt,e.retryMax); break;
    case 'stream_attempt': endModelActivity(); break; // desktop parity: a new attempt closes the previous window
    case 'turn_done': clearPendingPrompts(); finalizeMsg(); setRunning(false); endModelActivity(); autoSendGuidance(); refreshWorkspaceAfterTurn(); loadSessions();
      // Desktop behavior: a completed turn folds its tools+reasoning behind
      // the summary bar unless the user manually toggled it.
      { const tt = turnEls.get(currentTurn);
        if (tt && !tt.userOverride && tt.summary.style.display !== 'none') { tt.folded = true; applyTurnFold(tt); } }
      if(deliveryRecoveryActive&&e.outcome!=='final_readiness'&&!e.err)clearDeliveryCards();
      deliveryRecoveryActive=false;
      if(e.outcome==='final_readiness'){showDeliveryReadiness(e);}else if(e.outcome==='recovery_paused'){showNotice('⏸ '+__('recovery_paused'));}else if(e.err){log.appendChild(el('div','msg--error','✗ '+e.err));scrollDown();} fetchStatus(); fetchTodos(); refreshCheckpointAvailability(); break;
  }
};
es.onerror=()=>{
  if(es.readyState===EventSource.CONNECTING){setConnState('reconnecting');}
  else{setConnState('disconnected');}
};
}
__authReady.then(connectEvents).catch(error=>{setConnState('disconnected');showNotice(error instanceof Error?error.message:__('auth_failed'),'warn');});

// ── status polling ──
function fetchStatus(){
  fetch('/status').then(r=>r.json()).then(s=>{
    if(s.label)statusModel.textContent=s.label;
    if(welcomeModel)welcomeModel.textContent=s.label||'-';
    const cml=$('#composer-model-value');if(cml){const cur=modelsCache.find(m=>m.active);cml.textContent=(s.label||'-')+(cur&&cur.provider?' · '+providerLabel(cur.provider):'');}
    if(welcomeCwd){
      const cwd=String(s.workspaceRoot||s.cwd||'-');
      welcomeCwd.textContent=workspaceDisplayName(cwd);
      welcomeCwd.title=cwd;
      welcomeCwd.setAttribute('aria-label',__('workspace')+': '+cwd);
      workspaceRootPath=cwd==='-'?'':cwd;
      if(workspaceRootName){workspaceRootName.textContent=workspaceDisplayName(workspaceRootPath);workspaceRootName.title=workspaceRootPath;}
    }
    // Re-sync running from the server: after a history rebuild the SSE guard
    // skips turn_started/turn_done, so the server is the source of truth here.
    if(typeof s.running==='boolean'&&s.running!==running)setRunning(s.running);
    // sync mode UI without triggering POST (server is source of truth)
    planMode=!!s.plan;
    toolApprovalMode=s.toolApprovalMode || ((s.autoApproveTools??s.bypass)?'yolo':'ask');
    bypassMode=toolApprovalMode==='yolo';
    if(!bypassMode&&toolApprovalMode==='auto')yoloRestoreMode='auto';
    updateModeButtons();
    if(s.window){const pct=Math.min(100,Math.round(s.used/s.window*100));ctxFill.style.width=pct+'%';ctxFill.style.background=pct>85?'var(--warning)':pct>95?'var(--danger)':'var(--accent)';ctxUsed.textContent=fmtTok(s.used)+' tokens';ctxWindow.textContent=fmtTok(s.window)+' tokens';}
    // goal state
    goalText=(s.goal||'').trim();
    goalActive=goalText!==''&&(s.goalStatus||'')==='running';
    updateGoalUI();
    sessionCostQuote=s.sessionCostQuote||null;
    // sidebar metrics
    const cacheTotal=(s.cacheHit||0)+(s.cacheMiss||0);
    $('#sm-cache').textContent=cacheTotal>0?Math.round((s.cacheHit||0)/cacheTotal*100)+'%':'—';
    const lastPicked=usageSelectedCost(s.lastUsage);
    $('#sm-cost').textContent=lastPicked?.bucketed?__('multi_currency'):lastPicked?fmtMoney(lastPicked.amount,lastPicked.currency):'—';
    if(s.balance){$('#sm-balance').textContent=s.balance.display||'--';}
  }).catch(()=>{});
}
setInterval(fetchStatus,30000);

// ── history ──
function renderHistoryMessages(ms){
  if(!ms||ms.length===0){resetItems();hasVisibleHistory=false;showWelcome();updateActionAvailability();return false;}
  const visible = ms.some(m => {
    if(m.role==='user')return !!m.content;
    if(m.role==='assistant')return !!(m.content||m.reasoning||(m.toolCalls||[]).some(tc=>!hiddenTranscriptTool(tc.name)));
    if(m.role==='tool')return !hiddenTranscriptTool(m.toolName)&&!!(m.content||m.toolCallId||m.toolName);
    return false;
  });
  hasVisibleHistory=visible;
  if(!visible){resetItems();showWelcome();updateActionAvailability();return false;}
  hideWelcome();
  // Wholesale rebuild through the same items model the live stream uses, so
  // the reloaded transcript matches what was streamed. The welcome block
  // lives inside #log, so re-append it after clearing.
  resetItems();
  log.innerHTML='';
  if (welcome) log.appendChild(welcome);
  const resultById=new Map();
  ms.forEach(m=>{if(m.role==='tool'&&m.toolCallId&&!resultById.has(m.toolCallId))resultById.set(m.toolCallId,m);});
  const consumed=new Set();
  let seq=0;
  let histTurn=0; // /history has no turn ids; a user message opens a new turn
  ms.forEach(m=>{
    if(m.role==='system')return;
    if(m.role==='user'){
      if(m.content){const it={id:genItemId(),kind:'user',text:m.content,dom:null,turn:++histTurn};items.push(it);const d=renderItem(it);if(d)appendItem(it,d);}
      return;
    }
    if(m.role==='assistant'){
      const allCalls=m.toolCalls||[];
      allCalls.filter(tc=>hiddenTranscriptTool(tc.name)).forEach(tc=>{if(tc.id)consumed.add(tc.id);});
      const visibleCalls=allCalls.filter(tc=>!hiddenTranscriptTool(tc.name));
      if(!m.content&&!m.reasoning&&!visibleCalls.length)return;
      const it={id:genItemId(),kind:'assistant',text:m.content||'',reasoning:m.reasoning||'',tools:[],done:true,dom:null,turn:histTurn||++histTurn};
      visibleCalls.forEach(tc=>{
        const id=tc.id||'hist-tool-'+(seq++);
        const result=resultById.get(tc.id);
        const histTool={id,name:tc.name||'tool',args:String(tc.arguments||''),resolvedName:tc.resolvedName||'',capabilityId:tc.capabilityId||'',output:result?String(result.content||''):'',err:result?.err||'',status:result?'done':'running',readOnly:false,durationMs:result?Number(result.durationMs||0):0,added:Number(tc.added||0),removed:Number(tc.removed||0),subagentSummary:tc.subagentSummary||null,profile:tc.subagentSummary?{model:tc.subagentSummary.model||'',effort:tc.subagentSummary.effort||''}:null};
        if(tc.subagentSummary){const finalPreview=appendSubagentPreview('',histTool.output,SUBAGENT_PROGRESS_LIMITS.text);histTool.subagent={phase:tc.subagentSummary.status||'completed',reasoning:'',text:'',notice:'',truncated:finalPreview.truncated,startedAt:Number(tc.subagentSummary.startedAt||Date.now()),lastActivityAt:Number(tc.subagentSummary.endedAt||Date.now()),durationMs:Number(tc.subagentSummary.durationMs||0),finalOutput:finalPreview.value};}
        addToolAudit(histTool,capabilityAuditEntry(histTool));
        it.tools.push(histTool);
        if(tc.id)consumed.add(tc.id);
      });
      items.push(it);
      const d=renderItem(it);if(d)appendItem(it,d);
      return;
    }
    if(m.role==='tool'){
      if(m.toolCallId&&consumed.has(m.toolCallId))return;
      if(hiddenTranscriptTool(m.toolName))return;
      const id=m.toolCallId||'hist-tool-'+(seq++);
      const it={id,kind:'tool',name:m.toolName||'tool',args:'',output:m.content||'',err:m.err||'',status:m.err?'failed':'done',durationMs:Number(m.durationMs||0),dom:null,turn:histTurn||++histTurn};
      items.push(it);
      const d=renderItem(it);if(d)appendItem(it,d);
    }
  });
  // Desktop behavior: completed turns render collapsed — tools and reasoning
  // hide behind the count-style summary bar (unless the user toggled it).
  turnEls.forEach(t => {
    if (t.summary.style.display !== 'none' && !t.userOverride) { t.folded = true; applyTurnFold(t); }
  });
  // The next live turn must continue after the rebuilt history, or new
  // messages would collide with turn 1's container (and be hidden when that
  // turn folds). histTurn is the last turn number the rebuild produced.
  currentTurn = histTurn;
  syncQuestionNav();
  scrollDown(true); updateActionAvailability(); return true;
}
// reloadHistory fetches and rebuilds the transcript from /history, guarding
// the SSE stream while it runs (historyPending) so gap events are neither
// rendered-then-wiped nor lost silently — the snapshot covers them. Always
// clears the guard, even on failure, so live events resume.
function reloadHistory() {
  historyPending = true;
  fetch('/history').then(r => r.json()).then(msgs => {
    renderHistoryMessages(msgs);
    fetchTodos();
    refreshCheckpointAvailability();
  }).catch(() => {
    showNotice(__('error_loading'), 'warn');
  }).finally(() => {
    historyPending = false;
    fetchStatus(); // re-sync running/goal/context after the rebuild
  });
}
reloadHistory();

// ── session list ──
let sessionCount=0;
function sessionTitle(s){
  if(s.draft)return __('new_session_draft');
  const name=String(s.name||'').replace(/^.*\//,'').replace(/\.jsonl$/,'');
  return s.title||name.replace(/^\w+-/,'').replace(/T/,' ').replace(/[-_]/g,' ').slice(0,30);
}
function renderSessions(){
  const list=$('#session-list'); if(!list)return;
  const count=$('#session-count');
  const q=sessionFilter.trim().toLowerCase();
  const filtered=sessionsCache.filter(s=>{
    if(!q)return true;
    return [sessionTitle(s),s.name,s.path,String(s.turns||'')].some(v=>String(v||'').toLowerCase().includes(q));
  });
  sessionCount=sessionsCache.length;
  if(count)count.textContent=sessionCount?String(sessionCount):'';
  if(sessionsCache.length===0){list.innerHTML='<div style="padding:10px;color:var(--muted-2);font-size:12px">'+__('no_sessions')+'</div>';return;}
  if(filtered.length===0){list.innerHTML='<div style="padding:10px;color:var(--muted-2);font-size:12px">'+__('no_sessions')+'</div>';return;}
  list.innerHTML='';
  filtered.forEach(s=>{
    const item=el('div','session-item'+(s.current?' session-item--active':''));
    const title=sessionTitle(s);
    const meta=s.draft?__('session_draft'):(s.turns?s.turns+' turns':'');
    const del=(s.current||s.draft)?'':'<button type="button" class="session-del" data-name="'+escAttr(s.name)+'" title="'+escAttr(__('delete_session'))+'">&times;</button>';
    item.innerHTML='<svg class="session-item__icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg><div class="session-item__body"><div class="session-item__title">'+escHtml(title)+'</div><div class="session-item__meta">'+escHtml(meta)+'</div></div>'+del;
    item.onclick=ev=>{if(ev.target.closest('.session-del')||running||s.current)return;post('/resume',{path:s.path}).then(()=>{log.innerHTML='';log.appendChild(welcome);showWelcome();resetItems();hasVisibleHistory=false;checkpointCount=0;todosDismissed=false;resetCumulativeStats();loadSessions();updateActionAvailability();reloadHistory();fetchStatus();});};
    list.appendChild(item);
  });
}
let sessionsLoadSequence=0;
function loadSessions(){
  const sequence=++sessionsLoadSequence;
  fetch('/sessions').then(r=>r.json()).then(ss=>{
    if(sequence!==sessionsLoadSequence)return;
    sessionsCache=Array.isArray(ss)?ss:[];
    renderSessions();
  }).catch(()=>{if(sequence!==sessionsLoadSequence)return;const list=$('#session-list');if(list)list.innerHTML='<div style="padding:10px;color:var(--muted-2);font-size:12px">'+__('error_loading')+'</div>';});
}
loadSessions();

// ── branches panel ──
function branchValue(b, ...keys) {
  for(const k of keys){ if(b && b[k] != null && b[k] !== '') return b[k]; }
  return '';
}
function branchTitle(b) {
  return branchValue(b,'custom_title','CustomTitle','name','Name','topic_title','TopicTitle','id','ID') || 'branch';
}
function renderBranches(data) {
  const tree=$('#branches-tree'), list=$('#branches-list');
  if(!tree||!list)return;
  tree.textContent=data?.tree||'branches: none';
  const branches=Array.isArray(data?.branches)?data.branches:[];
  if(branches.length===0){
    list.innerHTML='<div class="empty-note">'+escHtml(__('no_branches'))+'</div>';
    return;
  }
  list.innerHTML='';
  branches.forEach(b=>{
    const id=String(branchValue(b,'id','ID'));
    const preview=String(branchValue(b,'preview','Preview'));
    const turns=branchValue(b,'turns','Turns');
    const model=String(branchValue(b,'model','Model'));
    const item=el('div','branch-item');
    const meta=[turns?turns+' turns':'',model,preview].filter(Boolean).join(' · ');
    item.innerHTML='<div><div class="branch-item__title">'+escHtml(branchTitle(b))+'</div><div class="branch-item__meta">'+escHtml(meta||id)+'</div></div><button class="branch-item__btn" data-branch-id="'+escAttr(id)+'">'+escHtml(__('switch'))+'</button>';
    list.appendChild(item);
  });
}
function openBranches() {
  if(!prepareOrdinaryOverlay())return;
  const modal=$('#branches-modal');
  if(!modal)return;
  $('#branches-tree').textContent=__('loading');
  $('#branches-list').innerHTML='';
  modal.style.display='flex';
  fetch('/branches').then(r=>r.json()).then(renderBranches).catch(()=>{
    $('#branches-tree').textContent=__('error_loading');
  });
}
function closeBranches(){const m=$('#branches-modal');if(m)m.style.display='none';}

// ── models panel ──
function renderModelsPanel(data) {
  const list=$('#models-list');
  if(!list)return;
  const models=Array.isArray(data?.models)?data.models:[];
  if(models.length===0){list.innerHTML='<div class="empty-note">'+escHtml(__('error_loading'))+'</div>';return;}
  list.innerHTML='';
  models.forEach(m=>{
    const active=!!m.active;
    const item=el('div','model-item'+(active?' model-item--active':''));
    item.innerHTML='<div><div class="model-item__title">'+escHtml(m.ref||'')+'</div><div class="model-item__meta">'+escHtml([m.kind,m.default?'default':''].filter(Boolean).join(' · '))+'</div></div>';
    if(active){
      item.innerHTML+='<span class="model-item__status model-item__status--active">'+escHtml(__('active'))+'</span>';
    }else{
      item.innerHTML+='<button class="branch-item__btn" data-model-ref="'+escAttr(m.ref||'')+'">'+escHtml(__('use_model'))+'</button>';
    }
    list.appendChild(item);
  });
}
function openModels() {
  if(!prepareOrdinaryOverlay())return;
  const modal=$('#models-modal');
  if(!modal)return;
  $('#models-list').innerHTML='<div class="empty-note">'+escHtml(__('loading'))+'</div>';
  modal.style.display='flex';
  fetch('/models').then(r=>r.json()).then(renderModelsPanel).catch(()=>{
    $('#models-list').innerHTML='<div class="empty-note">'+escHtml(__('error_loading'))+'</div>';
  });
}
function closeModels(){const m=$('#models-modal');if(m)m.style.display='none';}

// ── delete confirmation ──
function openDeleteSession(name) {
  if(!name||!prepareOrdinaryOverlay())return;
  pendingDeleteSession=name;
  const title=$('#delete-session-title');
  if(title)title.textContent=name;
  const modal=$('#delete-modal');
  if(modal)modal.style.display='flex';
}
function closeDeleteSession() {
  pendingDeleteSession=null;
  const modal=$('#delete-modal');
  if(modal)modal.style.display='none';
}

// ── image viewer ──
let imageViewer = null;
function openImageViewer(url) {
  if(!prepareOrdinaryOverlay())return;
  if (imageViewer) imageViewer.remove();
  imageViewer = el('div', 'image-viewer');
  const img = document.createElement('img');
  img.src = url;
  img.alt = '';
  imageViewer.appendChild(img);
  imageViewer.onclick = e => { if (e.target === imageViewer) closeImageViewer(); };
  document.body.appendChild(imageViewer);
}
function closeImageViewer(){if(imageViewer){imageViewer.remove();imageViewer=null;}}

// ── attachments (paste / drag-drop images into the composer) ──
function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result).split(',')[1] || '');
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
}
function insertIntoComposer(text) {
  const at = input.selectionStart ?? input.value.length;
  const prefix = input.value.slice(0, at);
  const suffix = input.value.slice(at);
  const sep = prefix && !/\n$/.test(prefix) ? '\n' : '';
  input.value = prefix + sep + text + '\n' + suffix;
  input.dispatchEvent(new Event('input'));
  input.focus();
}
async function attachFiles(files) {
  for (const f of files) {
    if (!f.type || !f.type.startsWith('image/')) continue;
    if (f.size > 10 << 20) continue;
    const b64 = await fileToBase64(f);
    const r = await post('/attach', { name: f.name || 'paste.png', data: b64 });
    if (r.ok) {
      const res = await r.json();
      if (res.path) insertIntoComposer('![' + (f.name || 'image') + '](' + res.path + ')');
    }
  }
}
input.addEventListener('paste', e => {
  const files = e.clipboardData && e.clipboardData.files;
  if (!files || !files.length) return;
  const imgs = Array.from(files).filter(f => f.type && f.type.startsWith('image/'));
  if (imgs.length) { e.preventDefault(); attachFiles(imgs); }
});
const composerBox = input.closest('.composer') || input.parentElement;
let dragDepth = 0;
composerBox.addEventListener('dragenter', e => { if (hasImageFiles(e)) { e.preventDefault(); dragDepth++; composerBox.classList.add('composer--drag'); } });
composerBox.addEventListener('dragover', e => { if (hasImageFiles(e)) { e.preventDefault(); } });
composerBox.addEventListener('dragleave', () => { if (--dragDepth <= 0) { dragDepth = 0; composerBox.classList.remove('composer--drag'); } });
composerBox.addEventListener('drop', e => {
  if (!hasImageFiles(e)) return;
  e.preventDefault();
  dragDepth = 0;
  composerBox.classList.remove('composer--drag');
  attachFiles(Array.from(e.dataTransfer.files));
});
function hasImageFiles(e) {
  return Array.from((e.dataTransfer && e.dataTransfer.files) || []).some(f => f.type && f.type.startsWith('image/'));
}

// ── input handling ──
async function syncModeBeforeSubmit(){
  await post('/plan',{on:planMode});
    await post('/tool-approval-mode',{mode:bypassMode?'yolo':toolApprovalMode});
  // If goalMode is on, we'll send as /goal command — no separate POST needed
}

// ── guidance queue (desktop composer-guidance parity) ──
// While a turn runs, submitted text joins this queue instead of starting a
// second turn. Each item can be steered into the active turn immediately
// (POST /steer); items left over when the turn ends are sent as ordinary
// follow-ups. Cancel folds the queue back into the draft.
function renderGuidanceShelf() {
  const shelf = $('#guidance-shelf');
  if (!shelf) return;
  const list = $('#guidance-list');
  if (guidanceQueue.length === 0) { shelf.style.display='none'; list.textContent=''; return; }
  shelf.style.display='';
  $('#guidance-count').textContent = __('guidance_count', 0, guidanceQueue.length);
  list.textContent='';
  const visible = guidanceExpanded ? guidanceQueue : guidanceQueue.slice(0, GUIDANCE_VISIBLE);
  for (const item of visible) {
    const row = el('div','composer-guidance-item');
    const icon = document.createElementNS('http://www.w3.org/2000/svg','svg');
    icon.setAttribute('viewBox','0 0 24 24'); icon.setAttribute('fill','none');
    icon.setAttribute('stroke','currentColor'); icon.setAttribute('stroke-width','2');
    icon.innerHTML = '<path d="m15 10 5 5-5 5"/><path d="M4 4v7a4 4 0 0 0 4 4h12"/>';
    icon.classList.add('composer-guidance-item__icon');
    row.appendChild(icon);
    row.appendChild(el('span','composer-guidance-item__text', item.text));
    const guide = el('button','composer-guidance-item__guide', __('guidance_mode'));
    guide.title = __('guidance_send');
    guide.disabled = !running || guidanceSendingId !== null;
    guide.onclick = () => void sendQueuedGuidance(item);
    row.appendChild(guide);
    const dismiss = el('button','composer-guidance-item__action','✕');
    dismiss.title = __('guidance_dismiss');
    dismiss.disabled = guidanceSendingId === item.id;
    dismiss.onclick = () => { guidanceQueue = guidanceQueue.filter(q => q.id !== item.id); renderGuidanceShelf(); };
    row.appendChild(dismiss);
    list.appendChild(row);
  }
  if (guidanceQueue.length > GUIDANCE_VISIBLE) {
    const more = el('button','composer-guidance-more',
      guidanceExpanded ? __('guidance_collapse') : __('guidance_remaining', 0, guidanceQueue.length - GUIDANCE_VISIBLE));
    more.setAttribute('aria-expanded', String(guidanceExpanded));
    more.onclick = () => { guidanceExpanded = !guidanceExpanded; renderGuidanceShelf(); };
    list.appendChild(more);
  }
}

function queueGuidance(text) {
  text = text.trim();
  if (!text) return;
  guidanceQueue.push({ id: guidanceNextId++, text });
  renderGuidanceShelf();
  input.value=''; input.style.height=''; closeSlashMenu();
}

// Steer one queued item into the active turn. On 409 the turn ended between
// our running check and the enqueue: if the turn is over, re-arm the auto-send
// so the item goes out as an ordinary follow-up; if the turn is still running,
// keep the item queued and warn.
async function sendQueuedGuidance(item) {
  if (guidanceSendingId !== null) return;
  const text = item.text.trim();
  if (!text) return;
  guidanceSendingId = item.id;
  renderGuidanceShelf();
  let rearmAutoSend = false;
  try {
    if (running) {
      const resp = await fetch('/steer', {method:'POST', headers:{'content-type':'application/json'}, body:JSON.stringify({text})});
      // The item may have left the queue while the fetch was in flight (a
      // stop folded it back into the draft): then the late result is moot.
      const stillQueued = guidanceQueue.some(q => q.id === item.id);
      if (resp.ok) {
        guidanceQueue = guidanceQueue.filter(q => q.id !== item.id);
        if (running && stillQueued) {
          // Steer accepted while the turn is still live. Mirror the backend
          // history render (serve.go historyMessages turns a queued steer
          // into a "↪ text" notice) so the user sees their guidance land in
          // the transcript while the turn runs. (202 only guarantees the
          // enqueue, not consumption: an un-consumed steer surfaces later as
          // the backend's unapplied notice.)
          const d = el('div','notice','↪ '+text);
          const it = { id: genItemId(), kind: 'notice', text: '↪ '+text, level: 'info', turn: currentTurn };
          items.push(it); appendItem(it, d); scrollDown();
        } else if (!running && stillQueued) {
          // The backend accepted the steer but the frontend already saw
          // turn_done: the agent's flush will record it as unapplied, so it
          // never reaches the model. Re-queue it and send it as an ordinary
          // follow-up (symmetric with the 409 path).
          guidanceQueue.push(item);
          rearmAutoSend = true;
        }
      } else if (stillQueued) {
        // 409: the turn ended between our running check and the enqueue.
        if (running) {
          showNotice(__('guidance_rejected'), 'warn');
        } else {
          rearmAutoSend = true;
        }
      }
    } else {
      appendUserMsg(text);
      await post('/submit',{input:text});
      guidanceQueue = guidanceQueue.filter(q => q.id !== item.id);
    }
  } finally {
    guidanceSendingId = null;
    renderGuidanceShelf();
    if (rearmAutoSend) autoSendGuidance();
  }
}

// After a turn ends, leftovers are sent as ordinary follow-ups (desktop:
// "queued guidance auto-sends once the turn finishes"). An item whose /steer
// is still in flight is left in the queue: its fetch result decides it, and
// sendQueuedGuidance re-arms this auto-send if the steer was rejected.
function autoSendGuidance() {
  const pending = guidanceQueue.filter(q => q.id !== guidanceSendingId);
  if (pending.length === 0) return;
  guidanceQueue = guidanceQueue.filter(q => q.id === guidanceSendingId);
  guidanceExpanded = false;
  renderGuidanceShelf();
  for (const item of pending) {
    appendUserMsg(item.text);
    post('/submit',{input:item.text}).then(r=>{
      if (!r.ok) { guidanceQueue.push(item); renderGuidanceShelf(); }
    }).catch(()=>{ guidanceQueue.push(item); renderGuidanceShelf(); });
  }
}

// Cancel restores queued guidance: stop means "stop acting", never "discard
// what I typed" (desktop composer-run-strip parity).
function foldGuidanceBackToDraft() {
  if (guidanceQueue.length === 0) return;
  const text = guidanceQueue.map(q => q.text).join('\n');
  guidanceQueue = [];
  guidanceExpanded = false;
  guidanceSendingId = null;
  renderGuidanceShelf();
  input.value = text;
  input.style.height='';
  input.focus();
}

async function send(){
  const v=input.value.trim(); if(!v)return;
  if(running){
    // Mid-turn: queue as guidance instead of starting a second turn.
    queueGuidance(v);
    return;
  }
  if(v==='/reload'){
    input.value='';input.style.height='';closeSlashMenu();
    showNotice(__('extensions_reloading'));
    const response=await post('/extensions/reload',{});
    if(response.ok){showNotice(__('extensions_reloaded'));fetchStatus();}
    else{showNotice((await response.text()).trim()||__('extensions_reload_failed'),'warn');}
    return;
  }
  await syncModeBeforeSubmit();
  let submitInput=v;
  if(goalMode && !v.startsWith('/goal')){
    // Send as /goal command for goal-draft mode
    submitInput='/goal '+v;
    goalMode=false;
    updateGoalUI();
  } else if(goalMode){
    // User typed /goal manually, exit goal-draft mode
    goalMode=false;
    updateGoalUI();
  }
  appendUserMsg(v);
  post('/submit',{input:submitInput}).then(r=>{if(r.ok&&(r.status===202||r.status===204)){fetchStatus();loadSessions();fetchEffort();}});
  input.value='';input.style.height='';closeSlashMenu();
}

input.addEventListener('input',()=>{input.style.height='auto';input.style.height=Math.min(input.scrollHeight,140)+'px';updateSlashMenu();});
input.addEventListener('keydown',e=>{
  // While an ask decision shelf is open it owns the keys (numbers/Enter/Esc
  // answer the question); the composer must not submit or type into the
  // transcript behind it.
  if(askActive){e.preventDefault();return;}
  // slash menu nav
  if(slashOpen){
    if(e.key==='ArrowDown'){e.preventDefault();if(slashArgMode){slashIndex=Math.min(slashIndex+1,slashArgItems.length-1);renderSlashArgs(input.value.slice(input.value.indexOf(' ')+1));}else{slashIndex=Math.min(slashIndex+1,slashFiltered.length-1);renderSlashMenu();}return;}
    if(e.key==='ArrowUp'){e.preventDefault();if(slashArgMode){slashIndex=Math.max(slashIndex-1,0);renderSlashArgs(input.value.slice(input.value.indexOf(' ')+1));}else{slashIndex=Math.max(slashIndex-1,0);renderSlashMenu();}return;}
    if(e.key==='Tab'||e.key==='Enter'){e.preventDefault();acceptSlash();return;}
    if(e.key==='Escape'){e.preventDefault();closeSlashMenu();return;}
  }
  // main keys
  if(e.key==='Enter'&&!e.shiftKey){e.preventDefault();send();return;}
  if(e.key==='Escape'){
    if(goalMode&&!running){goalMode=false;updateGoalUI();input.value='';closeSlashMenu();return;}
    if(running){foldGuidanceBackToDraft();post('/cancel');return;}
    // double-Esc for rewind
    if(input.value===''){if(escTimer){clearTimeout(escTimer);escTimer=null;openRewindPicker();}else{escTimer=setInterval(()=>escTimer=null,600);}return;}
  }
});

// shift+tab mode cycle
document.addEventListener('keydown',e=>{
  if(e.target===input&&e.key==='Tab'&&e.shiftKey){e.preventDefault();cycleMode();return;}
  if(e.target===input&&(e.key==='y'||e.key==='Y')&&(e.ctrlKey||e.metaKey)&&!e.altKey&&!e.shiftKey){e.preventDefault();toggleYolo();return;}
  if(e.key==='/'&&isPlainKey(e)){e.preventDefault();input.focus();return;}
});

// mode helpers — task mode (Direct/Plan/Goal, desktop collaborationMode).
const TASK_MODE_ICONS = {
  direct: '<path d="M12 19V5"/><path d="m5 12 7-7 7 7"/>',
  plan: '<path d="M8 6h13"/><path d="M8 12h13"/><path d="M8 18h13"/><path d="M3 6h.01"/><path d="M3 12h.01"/><path d="M3 18h.01"/>',
  goal: '<circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><path d="M12 2v4"/><path d="M12 18v4"/><path d="M2 12h4"/><path d="M18 12h4"/>',
};
function updateTaskModeUI(){
  const mode = goalActive || goalMode ? 'goal' : planMode ? 'plan' : 'direct';
  const btn = $('#btn-task-mode'), menu = $('#task-mode-menu');
  $('#task-mode-label').textContent = __('task_mode_'+mode);
  $('#task-mode-icon').innerHTML = TASK_MODE_ICONS[mode];
  btn.classList.toggle('task-mode__trigger--goal', mode==='goal');
  menu.querySelectorAll('.task-mode__item').forEach(it=>it.classList.toggle('is-active', it.dataset.mode===mode));
}
function updateModeButtons(){
  updateTaskModeUI();
  // approval modebar (desktop composer-modebar): ask / auto / yolo
  const cur = bypassMode ? 'yolo' : toolApprovalMode;
  const mb = $('#modebar');
  mb.dataset.mode = cur;
  mb.querySelectorAll('.composer-modebar__item').forEach(b => b.classList.toggle('is-active', b.dataset.mode === cur));
}
function updateGoalUI(){
  updateModeButtons();
  const bar=$('#goal-active-bar');
  if(goalActive&&goalText){
    bar.style.display='';
    $('#goal-chip-text').textContent=goalText;
  } else {
    bar.style.display='none';
  }
  input.placeholder=running?__('placeholder_running'):(goalMode?__('goal_placeholder'):__('placeholder'));
}
async function setToolApprovalMode(mode){toolApprovalMode=mode;bypassMode=mode==='yolo';updateModeButtons();await post('/tool-approval-mode',{mode});}
async function setPlan(on){planMode=on;updateModeButtons();await post('/plan',{on});}
async function cycleMode(){if(goalMode){goalMode=false;updateGoalUI();return;}await setPlan(!planMode);setTimeout(fetchStatus,200);}
async function toggleYolo(){if(bypassMode){const restore=yoloRestoreMode==='auto'?'auto':'ask';await setToolApprovalMode(restore);}else{yoloRestoreMode=toolApprovalMode==='auto'?'auto':'ask';await setToolApprovalMode('yolo');}setTimeout(fetchStatus,200);}
function toggleGoalMode(){
  if(goalActive){
    // Clear active goal
    post('/goal',{goal:''}).then(()=>{goalActive=false;goalText='';updateGoalUI();fetchStatus();});
    return;
  }
  if(goalMode){
    goalMode=false;
    updateGoalUI();
  } else {
    goalMode=true;
    updateGoalUI();
    input.focus();
  }
}

// ── todo panel ──
let todosState=[],todosDismissed=false;
function parseTodos(args){
  try{const a=JSON.parse(args);return Array.isArray(a.todos)?a.todos:[];}
  catch{return[];}
}
function todoStatusLabel(s){
  switch(String(s||'').trim()){
    case'completed':return '✓';
    case'in_progress':return '▶';
    default:return '○';
  }
}
function todoIsPhaseSignoff(ts,index){
  if(!ts[index]||Number(ts[index].level||0)!==0||!ts[index+1]||Number(ts[index+1].level||0)!==1)return false;
  for(let i=index+1;i<ts.length&&Number(ts[i].level||0)===1;i++)if(String(ts[i].status||'').trim()!=='completed')return false;
  return true;
}
function hasIncompleteTodos(ts){return ts.some(t=>String(t.status||'').trim()!=='completed');}
function renderTodoPanel(){
  const panel=$('#todo-panel'),list=$('#todos-list'),badge=$('#todos-badge');
  const summary=$('#todos-summary'),dismiss=$('#todos-dismiss'),title=$('#todos-title');
  if(!panel||todosDismissed){panel&&(panel.classList.remove('todos--visible'));return;}
  if(!todosState.length){panel.classList.remove('todos--visible');return;}
  const done=todosState.filter(t=>String(t.status||'').trim()==='completed').length;
  const total=todosState.length;
  const current=todosState.find(t=>String(t.status||'').trim()==='in_progress');
  const allDone=done===total;
  badge.textContent=done+'/'+total;
  summary.textContent=(current?.activeForm||current?.content||todosState[todosState.length-1]?.content||'').slice(0,60);
  dismiss.style.display=allDone?'':'none';
  if(allDone)panel.classList.add('todos--collapsed');
  list.innerHTML='';
  todosState.forEach((t,i)=>{
    const st=String(t.status||'').trim();
    const li=el('li','todos__item todos__item--'+st+(t.level?' todos__item--sub':''));
    const statusText=st==='in_progress'?(todoIsPhaseSignoff(todosState,i)?__('todo_phase_signoff'):__('todo_signable')):todoStatusLabel(st);
    li.innerHTML='<span class="todos__status todos__status--'+st+'">'+escHtml(statusText)+'</span><span class="todos__text">'+escHtml((st==='in_progress'&&t.activeForm)?t.activeForm:t.content)+'</span>';
    list.appendChild(li);
  });
  panel.classList.add('todos--visible');
}
function fetchTodos(){
  fetch('/todos').then(r=>r.json()).then(ts=>{
    if(Array.isArray(ts)){todosState=ts;renderTodoPanel();}
  }).catch(()=>{});
}
$('#todos-head').onclick=function(e){
  if(e.target.closest('.todos__dismiss')){todosDismissed=true;renderTodoPanel();return;}
  const panel=$('#todo-panel');panel.classList.toggle('todos--collapsed');
};

// toolbar buttons
btnSend.onclick=()=>void send();
btnStop.onclick=()=>{foldGuidanceBackToDraft();post('/cancel');};
// work mode (desktop runtime profile): full=Balanced / economy=Lightweight / delivery=Delivery
const WORKMODE_ICONS = {
  full: '<line x1="5" y1="12" x2="19" y2="12"/><line x1="5" y1="17" x2="19" y2="17"/>',
  economy: '<path d="m12 14 4-4"/><path d="M3.34 19a10 10 0 1 1 17.32 0"/>',
  delivery: '<path d="M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z"/><line x1="4" y1="22" x2="4" y2="15"/>',
};
const WORKMODE_KEYS = { full: 'work_mode_balanced', economy: 'work_mode_lightweight', delivery: 'work_mode_delivery' };
let workMode = 'full';
const workmodeMenu = $('#workmode-menu');
function updateWorkModeUI() {
  $('#workmode-label').textContent = __(WORKMODE_KEYS[workMode] || 'work_mode_balanced');
  $('#workmode-icon').innerHTML = WORKMODE_ICONS[workMode] || WORKMODE_ICONS.full;
  workmodeMenu.querySelectorAll('.workmode__item').forEach(it => it.classList.toggle('is-active', it.dataset.mode === workMode));
}
function fetchWorkMode() {
  fetch('/profile').then(r => r.json()).then(d => { if (d && d.mode) { workMode = d.mode; updateWorkModeUI(); } }).catch(() => {});
}
$('#btn-workmode').onclick=()=>{ if(running)return; workmodeMenu.style.display = workmodeMenu.style.display === 'none' ? '' : 'none'; };
workmodeMenu.querySelectorAll('.workmode__item').forEach(it=>{
  it.onclick=()=>{
    workmodeMenu.style.display='none';
    if (it.dataset.mode === workMode) return;
    post('/profile', { mode: it.dataset.mode }).then(() => fetchWorkMode());
  };
});
document.addEventListener('click', e => { if (workmodeMenu.style.display !== 'none' && !e.target.closest('.workmode')) workmodeMenu.style.display = 'none'; });
fetchWorkMode();
// task mode trigger
const taskModeMenu = $('#task-mode-menu');
$('#btn-task-mode').onclick=()=>{if(running)return;taskModeMenu.style.display=taskModeMenu.style.display==='none'?'':'none';};
taskModeMenu.querySelectorAll('.task-mode__item').forEach(it=>{
  it.onclick=async()=>{
    taskModeMenu.style.display='none';
    const m=it.dataset.mode;
    if(m==='direct'){if(goalMode)toggleGoalMode();await setPlan(false);}
    else if(m==='plan'){if(goalMode)toggleGoalMode();await setPlan(true);}
    else toggleGoalMode();
    setTimeout(fetchStatus,200);
  };
});
document.addEventListener('click',e=>{if(taskModeMenu.style.display!=='none'&&!e.target.closest('.task-mode'))taskModeMenu.style.display='none';});
// approval modebar items
$('#modebar').querySelectorAll('.composer-modebar__item').forEach(b=>{
  b.onclick=()=>{ if(waitingPrompt)return; if(b.dataset.mode==='yolo')void toggleYolo(); else void setToolApprovalMode(b.dataset.mode); };
});
$('#goal-chip').onclick=()=>toggleGoalMode();
$('#btn-new').onclick=()=>{if(running){showNotice(__('new_session_busy'),'warn');return;}post('/new').then(async response=>{if(!response.ok){showNotice((await response.text()).trim()||__('error_loading'),'warn');return;}log.innerHTML='';log.appendChild(welcome);showWelcome();resetItems();hasVisibleHistory=false;checkpointCount=0;todosState=[];todosDismissed=false;renderTodoPanel();resetCumulativeStats();sessionFilter='';const search=$('#session-search');if(search)search.value='';loadSessions();updateActionAvailability();fetchStatus();});};
// model switcher popover (desktop ModelSwitcher)
function providerLabel(p){
  switch(p){
    case 'deepseek':case 'deepseek-flash':case 'deepseek-pro':return __('provider_label_deepseek');
    default:return p;
  }
}
const modelswMenu = $('#modelsw-menu'), modelswList = $('#modelsw-list'), modelswSearch = $('#modelsw-search');
function renderModelGroups(data) {
  const models = Array.isArray(data?.models) ? data.models : [];
  modelsCache = models;
  const q = modelswSearch.value.trim().toLowerCase();
  const filtered = q ? models.filter(m => (m.model||'').toLowerCase().includes(q) || (m.provider||'').toLowerCase().includes(q)) : models;
  if (models.length === 0) { modelswList.innerHTML = '<div class="modelsw__empty">'+escHtml(__('error_loading'))+'</div>'; return; }
  if (filtered.length === 0) { modelswList.innerHTML = '<div class="modelsw__empty">'+escHtml(__('model_no_match'))+'</div>'; return; }
  const map = new Map();
  let curProviderLabel = '';
  for (const m of filtered) {
    if (m.active) curProviderLabel = providerLabel(m.provider) || '';
    const label = providerLabel(m.provider) || '?';
    const arr = map.get(label) || [];
    // De-duplicate same-named models inside a provider group: keep the
    // active/default entry, otherwise the first one (same behavior as the
    // desktop catalog where a provider maps to one display label).
    const dup = arr.find(x => (x.model || x.ref) === (m.model || m.ref));
    if (dup) {
      if ((m.active || m.default) && !(dup.active || dup.default)) {
        arr[arr.indexOf(dup)] = m;
      }
      continue;
    }
    arr.push(m);
    map.set(label, arr);
  }
  const groups = [...map.entries()].sort(([a],[b]) => a === curProviderLabel ? -1 : (b === curProviderLabel ? 1 : a.localeCompare(b)));
  modelswList.innerHTML = '';
  groups.forEach(([label, items]) => {
    const g = el('div', 'modelsw__group');
    g.appendChild(el('div', 'modelsw__group-label', label));
    items.forEach(m => {
      const b = el('button', 'modelsw__item' + (m.active ? ' modelsw__item--current' : ''));
      b.type = 'button';
      b.appendChild(el('span', 'modelsw__copy', m.model || m.ref || ''));
      if (m.active) b.appendChild(el('span', 'modelsw__check', '✓'));
      b.onclick = () => { closeModelsw(); post('/submit', { input: '/model ' + (m.ref || '') }).then(() => setTimeout(()=>{fetchStatus();fetchEffort();}, 300)); };
      g.appendChild(b);
    });
    modelswList.appendChild(g);
  });
}
function updateModelTrigger() {
  const cur = modelsCache.find(m => m.active);
  const cml = $('#composer-model-value');
  if (!cml) return;
  if (cur && cur.provider) cml.textContent = (cur.model || cur.ref || '-') + ' · ' + providerLabel(cur.provider);
}
function openModelsw() {
  modelswMenu.style.display = '';
  modelswSearch.value = '';
  modelswList.innerHTML = '<div class="modelsw__empty">'+escHtml(__('loading'))+'</div>';
  fetch('/models').then(r => r.json()).then(d => { renderModelGroups(d); updateModelTrigger(); }).catch(() => { modelswList.innerHTML = '<div class="modelsw__empty">'+escHtml(__('error_loading'))+'</div>'; });
  modelswSearch.focus();
}
function closeModelsw() { modelswMenu.style.display = 'none'; }
$('#btn-composer-model').onclick=()=>{ if(running)return; modelswMenu.style.display === 'none' ? openModelsw() : closeModelsw(); };
modelswSearch.addEventListener('input', () => renderModelGroups({ models: modelsCache }));
document.addEventListener('click', e => { if (modelswMenu.style.display !== 'none' && !e.target.closest('.modelsw')) closeModelsw(); });
// effort switcher (desktop EffortSwitcher)
const effortsw = $('#effortsw'), effortValue = $('#effort-value'), effortMenu = $('#effort-menu');
let effortState = null;
function renderEffortMenu() {
  effortMenu.innerHTML = '';
  (effortState?.levels || []).forEach(lv => {
    const b = el('button', 'effortsw__item' + (lv === effortState.current ? ' is-active' : ''));
    b.type = 'button';
    b.appendChild(el('span', 'effortsw__copy', lv));
    if (lv === effortState.current) b.appendChild(el('span', 'effortsw__check', '✓'));
    b.onclick = () => {
      effortMenu.style.display = 'none';
      if (lv === effortState.current) return;
      post('/effort', { level: lv }).then(() => fetchEffort());
    };
    effortMenu.appendChild(b);
  });
}
function fetchEffort() {
  fetch('/effort').then(r => r.json()).then(d => {
    effortState = d || {};
    const supported = !!(effortState.supported && Array.isArray(effortState.levels) && effortState.levels.length);
    effortsw.style.display = supported ? '' : 'none';
    if (supported) {
      const cur = effortState.current || 'auto';
      effortValue.textContent = cur;
      document.getElementById('btn-effort').classList.toggle('effortsw__trigger--explicit', cur !== 'auto');
      renderEffortMenu();
    }
  }).catch(() => {});
}
$('#btn-effort').onclick=()=>{ if(running)return; if(effortMenu.style.display==='none'){renderEffortMenu();effortMenu.style.display='';} else effortMenu.style.display='none'; };
document.addEventListener('click', e => { if (effortMenu.style.display !== 'none' && !e.target.closest('.effortsw')) effortMenu.style.display = 'none'; });
fetchEffort();
$('#btn-stats').onclick=()=>openStats();
$('#branches-modal-close').onclick=()=>closeBranches();
$('#branches-modal').onclick=e=>{if(e.target===e.currentTarget)closeBranches();};
$('#branches-list').onclick=e=>{
  const btn=e.target.closest('[data-branch-id]');
  if(!btn||running)return;
  const id=btn.dataset.branchId;
  if(!id)return;
  closeBranches();
  post('/submit',{input:'/switch '+id}).then(()=>{setTimeout(()=>{log.innerHTML='';log.appendChild(welcome);showWelcome();resetItems();loadSessions();reloadHistory();fetchStatus();},300);});
};
$('#models-modal-close').onclick=()=>closeModels();
$('#models-modal').onclick=e=>{if(e.target===e.currentTarget)closeModels();};
$('#models-list').onclick=e=>{
  const btn=e.target.closest('[data-model-ref]');
  if(!btn||running)return;
  const ref=btn.dataset.modelRef;
  if(!ref)return;
  closeModels();
  post('/submit',{input:'/model '+ref}).then(()=>setTimeout(()=>{fetchStatus();fetchEffort();},300));
};
$('#session-search').addEventListener('input',e=>{sessionFilter=e.target.value;renderSessions();});

// ── mobile sidebar toggle ──
const app=$('.app'),sidebar=$('.sidebar'),overlay=$('#sidebar-overlay'),menuBtn=$('#menu-btn'),sidebarCollapseBtn=$('#btn-sidebar-collapse');
let sidebarCollapsed=false;
try{sidebarCollapsed=localStorage.getItem('baize-sidebar-collapsed')==='true';}catch{}
function mobileLayout(){return window.matchMedia('(max-width:768px)').matches;}
function applySidebarCollapse(){
  app.classList.toggle('app--sidebar-collapsed',sidebarCollapsed);sidebar.classList.toggle('sidebar--collapsed',sidebarCollapsed);
  const expanded=mobileLayout()?sidebar.classList.contains('sidebar--open'):!sidebarCollapsed;
  sidebarCollapseBtn.setAttribute('aria-expanded',expanded?'true':'false');
  sidebarCollapseBtn.title=__(mobileLayout()||!sidebarCollapsed?'sidebar_collapse':'sidebar_expand');
  sidebarCollapseBtn.setAttribute('aria-label',sidebarCollapseBtn.title);
}
function openSidebar(){sidebar.classList.add('sidebar--open');overlay.classList.add('sidebar-overlay--visible');menuBtn.style.opacity='0';applySidebarCollapse();}
function closeSidebar(){sidebar.classList.remove('sidebar--open');overlay.classList.remove('sidebar-overlay--visible');menuBtn.style.opacity='';applySidebarCollapse();}
menuBtn.onclick=()=>openSidebar();
overlay.onclick=()=>closeSidebar();
sidebarCollapseBtn.onclick=()=>{
  if(mobileLayout()){closeSidebar();return;}
  sidebarCollapsed=!sidebarCollapsed;try{localStorage.setItem('baize-sidebar-collapsed',String(sidebarCollapsed));}catch{}applySidebarCollapse();
};
window.addEventListener('resize',applySidebarCollapse);applySidebarCollapse();

// ── workspace files ──
const workspacePanel=$('#workspace-panel'),workspaceTree=$('#workspace-tree'),workspaceSearchInput=$('#workspace-search');
const workspacePreviewContent=$('#workspace-preview-content'),workspacePreviewHead=$('#workspace-preview-head');
const workspaceRootName=$('#workspace-root-name'),workspaceHTMLToggle=$('#workspace-html-toggle');
let workspaceRootPath='',workspaceOpen=false,workspaceTreePath='',workspaceCurrentPreview=null,workspaceHTMLSource=false;
let workspaceTreeGeneration=0,workspacePreviewGeneration=0,workspaceSearchGeneration=0,workspaceSearchTimer=null;
function workspaceFormatBytes(size){
  const n=Number(size)||0;if(n<1024)return n+' B';if(n<1048576)return(n/1024).toFixed(n<10240?1:0)+' KiB';return(n/1048576).toFixed(1)+' MiB';
}
function workspaceState(text){workspaceTree.innerHTML='';workspaceTree.appendChild(el('div','workspace-tree__state',text));}
async function workspaceJSON(path){
  const response=await fetch(path);
  if(!response.ok){const error=new Error('HTTP '+response.status);error.status=response.status;throw error;}
  return response.json();
}
function openWorkspacePanel(){
  if(!ordinaryOverlayAllowed())return false;
  closeSettings({restoreFocus:false});closeOrdinaryModals();
  if(!workspaceOpen){workspaceOpen=true;app.classList.add('app--workspace-open');workspacePanel.setAttribute('aria-hidden','false');$('#btn-workspace').classList.add('sidebar__item--active');loadWorkspaceTree(workspaceTreePath);}
  if(mobileLayout())closeSidebar();
  return true;
}
function closeWorkspacePanel({restoreFocus=false}={}){const wasOpen=workspaceOpen;workspaceOpen=false;app.classList.remove('app--workspace-open');workspacePanel.setAttribute('aria-hidden','true');$('#btn-workspace').classList.remove('sidebar__item--active');if(wasOpen&&restoreFocus)$('#btn-workspace').focus();}
async function loadWorkspaceTree(path='',throwOnError=false){
  const generation=++workspaceTreeGeneration;workspaceTreePath=path;workspaceState(__('loading'));
  try{
    const data=await workspaceJSON('/workspace/entries?path='+encodeURIComponent(path));if(generation!==workspaceTreeGeneration)return false;
    workspaceTree.innerHTML='';
    if(path){const up=el('button','workspace-tree__row');up.type='button';up.innerHTML='<span class="workspace-tree__twisty">←</span><span class="workspace-tree__icon">⌂</span><span class="workspace-tree__name">'+escHtml(path)+'</span>';up.onclick=()=>loadWorkspaceTree('');workspaceTree.appendChild(up);}
    renderWorkspaceEntries(workspaceTree,Array.isArray(data.entries)?data.entries:[]);
    if(!data.entries?.length)workspaceState(__('workspace_empty'));
    return true;
  }catch(error){if(generation===workspaceTreeGeneration)workspaceState(__('workspace_load_error'));if(throwOnError)throw error;return false;}
}
function renderWorkspaceEntries(container,entries){
  entries.forEach(entry=>{
    const row=el('button','workspace-tree__row'+(workspaceCurrentPreview?.path===entry.path?' workspace-tree__row--active':''));row.type='button';row.title=entry.path;row.setAttribute('role','treeitem');
    const twist=el('span','workspace-tree__twisty',entry.isDir?'›':'');const icon=el('span','workspace-tree__icon',entry.isDir?'▰':'▤');const name=el('span','workspace-tree__name',entry.name);
    row.append(twist,icon,name);container.appendChild(row);
    if(entry.isDir){row.setAttribute('aria-expanded','false');row.onclick=()=>toggleWorkspaceDirectory(row,entry.path,twist);}else row.onclick=()=>loadWorkspacePreview(entry.path);
  });
}
async function toggleWorkspaceDirectory(row,path,twist){
  let children=row.nextElementSibling;
  if(children?.classList.contains('workspace-tree__children')){const hidden=children.hidden;children.hidden=!hidden;row.setAttribute('aria-expanded',hidden?'true':'false');twist.textContent=hidden?'⌄':'›';return;}
  children=el('div','workspace-tree__children');children.appendChild(el('div','workspace-tree__state',__('loading')));row.after(children);row.setAttribute('aria-expanded','true');twist.textContent='⌄';
  try{const data=await workspaceJSON('/workspace/entries?path='+encodeURIComponent(path));if(!children.isConnected)return;children.innerHTML='';renderWorkspaceEntries(children,Array.isArray(data.entries)?data.entries:[]);if(!data.entries?.length)children.appendChild(el('div','workspace-tree__state',__('workspace_empty')));}catch{children.innerHTML='';children.appendChild(el('div','workspace-tree__state',__('workspace_load_error')));}
}
async function searchWorkspace(query){
  const generation=++workspaceSearchGeneration;if(query.length<2){loadWorkspaceTree('');return;}workspaceState(__('loading'));
  try{const data=await workspaceJSON('/workspace/search?q='+encodeURIComponent(query));if(generation!==workspaceSearchGeneration)return;workspaceTree.innerHTML='';const entries=Array.isArray(data.entries)?data.entries:[];entries.forEach(entry=>{const row=el('button','workspace-tree__row');row.type='button';row.title=entry.path;row.append(el('span','workspace-tree__twisty',''),el('span','workspace-tree__icon',entry.isDir?'▰':'▤'),el('span','workspace-tree__name',entry.path));row.onclick=()=>{if(entry.isDir){workspaceSearchInput.value='';loadWorkspaceTree(entry.path);}else loadWorkspacePreview(entry.path);};workspaceTree.appendChild(row);});if(!entries.length)workspaceState(__('workspace_no_results'));}catch{if(generation===workspaceSearchGeneration)workspaceState(__('workspace_load_error'));}
}
function workspacePreviewNote(text){workspacePreviewContent.appendChild(el('div','workspace-preview__note',text));}
function workspaceCodeView(body,path){
  const pre=el('pre','workspace-preview__code'),code=el('code');code.textContent=body;const ext=(String(path).match(/\.([^.\/]+)$/)||[])[1]||'';
  const aliases={js:'javascript',ts:'typescript',jsx:'javascript',tsx:'typescript',yml:'yaml',ps1:'powershell',sh:'bash',md:'markdown'};const language=aliases[ext.toLowerCase()]||ext.toLowerCase();
  if(language&&hljs.getLanguage(language))code.className='language-'+language;pre.appendChild(code);workspacePreviewContent.appendChild(pre);try{hljs.highlightElement(code);}catch{}
}
function workspaceChartScriptURL(value){
  try{
    const url=new URL(value);
    if(url.origin!=='https://cdn.jsdelivr.net'||url.search||url.hash)return'';
    if(!/^\/npm\/echarts@\d+\.\d+\.\d+\/dist\/echarts\.min\.js$/.test(url.pathname))return'';
    return url.href;
  }catch{return'';}
}
function safeWorkspaceHTML(raw){
  const source=new DOMParser().parseFromString(raw,'text/html'),sourceScripts=Array.from(source.scripts);
  const externalScripts=sourceScripts.filter(script=>script.src);
  const chartScripts=externalScripts.map(script=>workspaceChartScriptURL(script.getAttribute('src')||''));
  const interactive=externalScripts.length>0&&chartScripts.every(Boolean)&&sourceScripts.every(script=>!script.type||/^(?:text\/javascript|application\/javascript)$/i.test(script.type));
  const clean=DOMPurify.sanitize(raw,{WHOLE_DOCUMENT:true,SANITIZE_DOM:!interactive,USE_PROFILES:{html:true,svg:true,svgFilters:true},FORBID_TAGS:['script','iframe','object','embed','form','input','button','meta','link','base'],FORBID_ATTR:['srcset']});
  const doc=new DOMParser().parseFromString(clean,'text/html');
  doc.querySelectorAll('[src],[href],[xlink\\:href],[action],[formaction]').forEach(node=>{
    for(const attr of ['src','href','xlink:href','action','formaction']){if(!node.hasAttribute(attr))continue;const value=(node.getAttribute(attr)||'').trim();if((attr==='src'&&/^data:image\//i.test(value))||((attr==='href'||attr==='xlink:href')&&value.startsWith('#')))continue;node.removeAttribute(attr);}
    node.removeAttribute('target');
  });
  const scriptPolicy=interactive?"'unsafe-inline' "+chartScripts.join(' '):"'none'";
  const meta=doc.createElement('meta');meta.httpEquiv='Content-Security-Policy';meta.content="default-src 'none'; img-src data:; style-src 'unsafe-inline'; font-src data:; script-src "+scriptPolicy+"; connect-src 'none'; frame-src 'none'; object-src 'none'; form-action 'none'; base-uri 'none'";doc.head.prepend(meta);
  if(interactive){
    sourceScripts.forEach((sourceScript,index)=>{
      const script=doc.createElement('script');
      if(sourceScript.src)script.src=chartScripts[externalScripts.indexOf(sourceScript)];else script.textContent=sourceScript.textContent;
      doc.body.appendChild(script);
    });
  }
  return{html:'<!doctype html>'+doc.documentElement.outerHTML,interactive};
}
function renderWorkspaceHTML(preview){
  if(workspaceHTMLSource){workspaceCodeView(preview.body,preview.path);workspaceHTMLToggle.textContent=__('workspace_report');return;}
  workspaceHTMLToggle.textContent=__('workspace_source');if(preview.truncated){workspacePreviewNote(__('workspace_truncated'));workspaceCodeView(preview.body,preview.path);return;}
  const rendered=safeWorkspaceHTML(preview.body),frame=el('iframe','workspace-preview__frame');frame.setAttribute('sandbox',rendered.interactive?'allow-scripts':'');frame.setAttribute('referrerpolicy','no-referrer');frame.title=preview.name;frame.srcdoc=rendered.html;workspacePreviewContent.appendChild(frame);
}
let workspacePDFModulePromise=null,workspacePDFState=null;
function disposeWorkspacePDF(){
  const state=workspacePDFState;workspacePDFState=null;if(!state)return;
  state.disposed=true;(state.pageViews||[]).forEach(view=>{view.generation++;if(view.renderTask)try{view.renderTask.cancel();}catch{}if(view.pdfPage)try{view.pdfPage.cleanup();}catch{}});
  if(state.resizeObserver)state.resizeObserver.disconnect();
  if(state.resizeFrame)cancelAnimationFrame(state.resizeFrame);
  if(state.scrollFrame)cancelAnimationFrame(state.scrollFrame);
  if(state.loadingTask)state.loadingTask.destroy().catch(()=>{});
}
function workspacePDFModule(){
  if(!workspacePDFModulePromise)workspacePDFModulePromise=import('/assets/pdfjs/pdf.mjs').then(pdfjs=>{pdfjs.GlobalWorkerOptions.workerSrc='/assets/pdfjs/pdf.worker.mjs';return pdfjs;});
  return workspacePDFModulePromise;
}
function workspacePDFAction(label,className){const node=el('button','workspace-pdf__button '+className,label);node.type='button';return node;}
function workspacePDFLink(label,className,preview,download){
  const node=el('a','workspace-pdf__button '+className,label);node.href=preview.contentUrl;
  if(download)node.download=preview.name;else{node.target='_blank';node.rel='noopener noreferrer';}
  return node;
}
function workspacePDFMessage(state,text,isError){
  state.status.textContent=text;state.status.classList.toggle('workspace-pdf__status--error',!!isError);state.status.hidden=false;
}
function updateWorkspacePDFControls(state){
  const ready=!!state.document;state.previous.disabled=!ready||state.page<=1;state.next.disabled=!ready||state.page>=state.pages;
  state.zoomOut.disabled=!ready;state.zoomIn.disabled=!ready;state.fit.disabled=!ready;
  state.pageInput.disabled=!ready;state.pageInput.max=String(Math.max(1,state.pages));state.pageInput.value=String(state.page);state.pageTotal.textContent=String(state.pages||'—');
  state.fit.classList.toggle('workspace-pdf__button--active',state.fitWidth);
}
function workspacePDFViewScale(state,view){
  if(!state.fitWidth)return state.scale;
  const width=view.naturalWidth||state.referenceWidth||612,available=Math.max(120,state.viewport.clientWidth-32);
  return Math.max(.25,Math.min(4,available/width));
}
function layoutWorkspacePDFView(state,view){
  const width=view.naturalWidth||state.referenceWidth||612,height=view.naturalHeight||state.referenceHeight||792,scale=workspacePDFViewScale(state,view);
  view.node.style.width=(width*scale)+'px';view.node.style.height=(height*scale)+'px';
}
function releaseWorkspacePDFView(view){
  view.generation++;if(view.renderTask)try{view.renderTask.cancel();}catch{}view.renderTask=null;view.renderKey='';view.canvas.hidden=true;view.canvas.width=1;view.canvas.height=1;
  if(view.pdfPage)try{view.pdfPage.cleanup();}catch{}view.pdfPage=null;
}
async function renderWorkspacePDFView(state,view){
  if(workspacePDFState!==state||state.disposed||!state.document)return;
  const scale=workspacePDFViewScale(state,view),renderKey=scale.toFixed(4)+'@'+Math.max(1,window.devicePixelRatio||1);if(view.renderKey===renderKey&&!view.canvas.hidden)return;
  const generation=++view.generation;if(view.renderTask)try{view.renderTask.cancel();}catch{}
  try{
    const page=await state.document.getPage(view.page);if(workspacePDFState!==state||state.disposed||generation!==view.generation)return;
    view.pdfPage=page;const natural=page.getViewport({scale:1});view.naturalWidth=natural.width;view.naturalHeight=natural.height;layoutWorkspacePDFView(state,view);
    const actualScale=workspacePDFViewScale(state,view),viewport=page.getViewport({scale:actualScale}),pixelRatio=Math.max(1,window.devicePixelRatio||1),canvas=view.canvas;
    canvas.width=Math.max(1,Math.floor(viewport.width*pixelRatio));canvas.height=Math.max(1,Math.floor(viewport.height*pixelRatio));canvas.style.width=viewport.width+'px';canvas.style.height=viewport.height+'px';
    canvas.hidden=false;view.message.hidden=true;const context=canvas.getContext('2d',{alpha:false});view.renderTask=page.render({canvasContext:context,viewport,transform:pixelRatio===1?null:[pixelRatio,0,0,pixelRatio,0,0]});
    await view.renderTask.promise;if(workspacePDFState!==state||state.disposed||generation!==view.generation)return;
    view.renderKey=actualScale.toFixed(4)+'@'+pixelRatio;if(view.page===state.page)state.status.hidden=true;
  }catch(error){
    if(error?.name==='RenderingCancelledException'||workspacePDFState!==state||state.disposed||generation!==view.generation)return;
    view.canvas.hidden=true;view.message.textContent=__('workspace_pdf_error');view.message.hidden=false;if(view.page===state.page)workspacePDFMessage(state,__('workspace_pdf_error'),true);
  }finally{if(workspacePDFState===state&&generation===view.generation)view.renderTask=null;}
}
function syncWorkspacePDFViews(state){
  if(!state.document)return;state.pageViews.forEach(view=>{layoutWorkspacePDFView(state,view);if(Math.abs(view.page-state.page)<=1)renderWorkspacePDFView(state,view);else releaseWorkspacePDFView(view);});updateWorkspacePDFControls(state);
}
function scrollWorkspacePDFToPage(state,page,behavior){
  page=Math.max(1,Math.min(state.pages,page));state.page=page;syncWorkspacePDFViews(state);const view=state.pageViews[page-1];if(view)state.viewport.scrollTo({top:Math.max(0,view.node.offsetTop-16),behavior:behavior||'auto'});
}
function updateWorkspacePDFPageFromScroll(state){
  state.scrollFrame=0;if(!state.document||state.scrollingToPage)return;const focus=state.viewport.scrollTop+Math.max(1,state.viewport.clientHeight*.3);let current=state.pageViews[0];
  for(const view of state.pageViews){if(view.node.offsetTop<=focus)current=view;else break;}
  if(current&&current.page!==state.page){state.page=current.page;syncWorkspacePDFViews(state);}
}
async function renderWorkspacePDF(preview){
  const root=el('div','workspace-pdf'),toolbar=el('div','workspace-pdf__toolbar'),viewport=el('div','workspace-pdf__viewport');
  const pagesNode=el('div','workspace-pdf__pages'),status=el('div','workspace-pdf__status',__('workspace_pdf_loading'));viewport.append(pagesNode,status);root.append(toolbar,viewport);workspacePreviewContent.appendChild(root);
  const previous=workspacePDFAction('←','workspace-pdf__previous'),next=workspacePDFAction('→','workspace-pdf__next');previous.setAttribute('aria-label',__('workspace_pdf_previous'));previous.title=__('workspace_pdf_previous');next.setAttribute('aria-label',__('workspace_pdf_next'));next.title=__('workspace_pdf_next');
  const pageLabel=el('label','workspace-pdf__page'),pageText=el('span','',__('workspace_pdf_page')),pageInput=document.createElement('input'),pageSeparator=el('span','workspace-pdf__page-separator','/'),pageTotal=el('span','workspace-pdf__page-total','—');
  pageInput.type='number';pageInput.min='1';pageInput.inputMode='numeric';pageInput.setAttribute('aria-label',__('workspace_pdf_page'));pageLabel.append(pageText,pageInput,pageSeparator,pageTotal);
  const zoomOut=workspacePDFAction('−','workspace-pdf__zoom-out'),zoomIn=workspacePDFAction('+','workspace-pdf__zoom-in'),fit=workspacePDFAction(__('workspace_pdf_fit_width'),'workspace-pdf__fit');zoomOut.setAttribute('aria-label',__('workspace_pdf_zoom_out'));zoomOut.title=__('workspace_pdf_zoom_out');zoomIn.setAttribute('aria-label',__('workspace_pdf_zoom_in'));zoomIn.title=__('workspace_pdf_zoom_in');fit.title=__('workspace_pdf_fit_width');
  const spacer=el('span','workspace-pdf__spacer'),open=workspacePDFLink(__('workspace_pdf_open'),'workspace-pdf__open',preview,false),download=workspacePDFLink(__('workspace_pdf_download'),'workspace-pdf__download',preview,true);toolbar.append(previous,next,pageLabel,zoomOut,zoomIn,fit,spacer,open,download);
  const state={preview,root,viewport,pagesNode,status,previous,next,pageInput,pageTotal,zoomOut,zoomIn,fit,page:1,pages:0,scale:1,fitWidth:true,document:null,loadingTask:null,pageViews:[],referenceWidth:612,referenceHeight:792,resizeObserver:null,resizeFrame:0,scrollFrame:0,passwordProtected:false,disposed:false};workspacePDFState=state;updateWorkspacePDFControls(state);
  previous.onclick=()=>{if(state.page>1)scrollWorkspacePDFToPage(state,state.page-1,'smooth');};next.onclick=()=>{if(state.page<state.pages)scrollWorkspacePDFToPage(state,state.page+1,'smooth');};
  pageInput.onchange=()=>{const page=Math.max(1,Math.min(state.pages,Number.parseInt(pageInput.value,10)||state.page));scrollWorkspacePDFToPage(state,page,'smooth');};
  zoomOut.onclick=()=>{const current=state.pageViews[state.page-1];if(state.fitWidth)state.scale=workspacePDFViewScale(state,current||{});state.fitWidth=false;state.scale=Math.max(.25,state.scale-.2);scrollWorkspacePDFToPage(state,state.page);};zoomIn.onclick=()=>{const current=state.pageViews[state.page-1];if(state.fitWidth)state.scale=workspacePDFViewScale(state,current||{});state.fitWidth=false;state.scale=Math.min(4,state.scale+.2);scrollWorkspacePDFToPage(state,state.page);};fit.onclick=()=>{state.fitWidth=true;scrollWorkspacePDFToPage(state,state.page);};
  viewport.addEventListener('scroll',()=>{if(!state.scrollFrame)state.scrollFrame=requestAnimationFrame(()=>updateWorkspacePDFPageFromScroll(state));},{passive:true});
  state.resizeObserver=new ResizeObserver(()=>{if(!state.fitWidth||!state.document)return;if(state.resizeFrame)cancelAnimationFrame(state.resizeFrame);state.resizeFrame=requestAnimationFrame(()=>{state.resizeFrame=0;scrollWorkspacePDFToPage(state,state.page);});});state.resizeObserver.observe(viewport);
  try{
    const pdfjs=await workspacePDFModule();if(workspacePDFState!==state)return;
    state.loadingTask=pdfjs.getDocument({url:preview.contentUrl,cMapUrl:'/assets/pdfjs/cmaps/',cMapPacked:true,standardFontDataUrl:'/assets/pdfjs/standard_fonts/',wasmUrl:'/assets/pdfjs/wasm/',iccUrl:'/assets/pdfjs/iccs/',isEvalSupported:false,enableXfa:false});
    state.loadingTask.onPassword=()=>{state.passwordProtected=true;workspacePDFMessage(state,__('workspace_pdf_protected'),true);state.loadingTask.destroy().catch(()=>{});};
    state.document=await state.loadingTask.promise;if(workspacePDFState!==state)return;state.pages=state.document.numPages;
    const firstPage=await state.document.getPage(1);if(workspacePDFState!==state)return;const natural=firstPage.getViewport({scale:1});state.referenceWidth=natural.width;state.referenceHeight=natural.height;try{firstPage.cleanup();}catch{}
    for(let page=1;page<=state.pages;page++){const node=el('section','workspace-pdf__page-view'),canvas=document.createElement('canvas'),message=el('span','workspace-pdf__page-message',__('workspace_pdf_rendering')),number=el('span','workspace-pdf__page-number',String(page));canvas.className='workspace-pdf__canvas';canvas.hidden=true;canvas.setAttribute('aria-label',preview.name+' '+page);node.append(canvas,message,number);pagesNode.appendChild(node);state.pageViews.push({page,node,canvas,message,pdfPage:null,renderTask:null,renderKey:'',generation:0,naturalWidth:0,naturalHeight:0});}
    updateWorkspacePDFControls(state);syncWorkspacePDFViews(state);
  }catch(error){
    if(workspacePDFState!==state)return;workspacePDFMessage(state,state.passwordProtected||error?.name==='PasswordException'?__('workspace_pdf_protected'):__('workspace_pdf_error'),true);updateWorkspacePDFControls(state);
  }
}
function renderWorkspacePreview(preview){
  disposeWorkspacePDF();workspacePreviewContent.innerHTML='';workspacePreviewContent.classList.toggle('workspace-preview__content--pdf',preview.kind==='pdf');workspacePreviewHead.style.display='flex';$('#workspace-preview-name').textContent=preview.name;$('#workspace-preview-name').title=preview.path;$('#workspace-preview-meta').textContent=preview.path+' · '+workspaceFormatBytes(preview.size);workspaceHTMLToggle.style.display=preview.kind==='html'?'':'none';
  if(preview.truncated&&preview.kind!=='html')workspacePreviewNote(__('workspace_truncated'));
  if(preview.kind==='markdown'){const body=el('div','md-sections workspace-preview__markdown');body.innerHTML=renderMarkdown(preview.body);fixImageSrcs(body,preview.path);highlightBlocks(body);workspacePreviewContent.appendChild(body);return;}
  if(preview.kind==='code'||preview.kind==='text'){workspaceCodeView(preview.body,preview.path);return;}
  if(preview.kind==='html'){renderWorkspaceHTML(preview);return;}
  if(preview.kind==='image'){const wrap=el('div','workspace-preview__media'),image=el('img');image.src=preview.contentUrl;image.alt=preview.name;image.onclick=()=>openImageViewer(preview.contentUrl);wrap.appendChild(image);workspacePreviewContent.appendChild(wrap);return;}
  if(preview.kind==='pdf'){renderWorkspacePDF(preview);return;}
  workspacePreviewContent.appendChild(el('div','workspace-preview__binary',__('workspace_binary')));
}
async function loadWorkspacePreview(path){
  if(!openWorkspacePanel())return false;
  const generation=++workspacePreviewGeneration;disposeWorkspacePDF();workspacePreviewContent.classList.remove('workspace-preview__content--pdf');workspacePreviewContent.innerHTML='';workspacePreviewContent.appendChild(el('div','workspace-empty',__('loading')));
  try{
    const preview=await workspaceJSON('/workspace/preview?path='+encodeURIComponent(path));if(generation!==workspacePreviewGeneration)return false;workspaceCurrentPreview=preview;workspaceHTMLSource=false;renderWorkspacePreview(preview);workspaceTree.querySelectorAll('.workspace-tree__row').forEach(row=>row.classList.toggle('workspace-tree__row--active',row.title===preview.path));return true;
  }catch(error){
    if(generation!==workspacePreviewGeneration)return false;
    if(error?.status===404){
      try{if(await loadWorkspaceTree(path,true)){workspaceCurrentPreview=null;workspacePreviewHead.style.display='none';workspacePreviewContent.innerHTML='';workspacePreviewContent.appendChild(el('div','workspace-empty',__('workspace_select_file')));return true;}}catch{}
    }
    workspacePreviewHead.style.display='none';workspacePreviewContent.innerHTML='';workspacePreviewContent.appendChild(el('div','workspace-empty',__('workspace_preview_error')+(error?.status?' (HTTP '+error.status+')':'')));return false;
  }
}
function openWorkspacePath(path){if(!ordinaryOverlayAllowed())return;workspaceSearchInput.value='';void loadWorkspacePreview(path);}
function openWorkspaceFile(path){openWorkspacePath(path);}
function refreshWorkspaceAfterTurn(){if(!workspaceOpen)return;loadWorkspaceTree(workspaceTreePath);if(workspaceCurrentPreview)loadWorkspacePreview(workspaceCurrentPreview.path);}
$('#btn-workspace').onclick=()=>workspaceOpen?closeWorkspacePanel():openWorkspacePanel();
$('#workspace-close').onclick=()=>closeWorkspacePanel({restoreFocus:true});
$('#workspace-refresh').onclick=()=>{loadWorkspaceTree(workspaceTreePath);if(workspaceCurrentPreview)loadWorkspacePreview(workspaceCurrentPreview.path);};
workspaceSearchInput.addEventListener('input',()=>{clearTimeout(workspaceSearchTimer);const query=workspaceSearchInput.value.trim();workspaceSearchTimer=setTimeout(()=>searchWorkspace(query),180);});
workspaceHTMLToggle.onclick=()=>{if(!workspaceCurrentPreview)return;workspaceHTMLSource=!workspaceHTMLSource;workspacePreviewContent.innerHTML='';renderWorkspaceHTML(workspaceCurrentPreview);};
$('#workspace-copy-path').onclick=async()=>{if(!workspaceCurrentPreview)return;const button=$('#workspace-copy-path');const ok=await copyText(workspaceCurrentPreview.path);button.title=ok?__('tool_copied'):__('workspace_copy_path');setTimeout(()=>{button.title=__('workspace_copy_path');},1200);};
const WORKSPACE_WIDTH_DEFAULT=720,WORKSPACE_WIDTH_MIN=480,WORKSPACE_WIDTH_MAX=1200,workspaceResizer=$('#workspace-resizer');
function workspaceWidthBounds(){return{min:WORKSPACE_WIDTH_MIN,max:Math.max(WORKSPACE_WIDTH_MIN,Math.min(WORKSPACE_WIDTH_MAX,window.innerWidth*.8))};}
function clampWorkspaceWidth(value){const bounds=workspaceWidthBounds();return Math.max(bounds.min,Math.min(bounds.max,Number(value)||WORKSPACE_WIDTH_DEFAULT));}
function applyWorkspaceWidth(value,persist=false){const width=clampWorkspaceWidth(value),bounds=workspaceWidthBounds();app.style.setProperty('--workspace-width',width+'px');workspaceResizer.setAttribute('aria-valuemin',String(bounds.min));workspaceResizer.setAttribute('aria-valuemax',String(Math.round(bounds.max)));workspaceResizer.setAttribute('aria-valuenow',String(Math.round(width)));if(persist)try{localStorage.setItem('baize-workspace-width',String(width));}catch{}return width;}
const storedWorkspaceWidth=Number.parseFloat((()=>{try{return localStorage.getItem('baize-workspace-width')||'';}catch{return '';}})());
applyWorkspaceWidth(Number.isFinite(storedWorkspaceWidth)?storedWorkspaceWidth:WORKSPACE_WIDTH_DEFAULT);
$('#workspace-resizer').addEventListener('pointerdown',event=>{
  if(mobileLayout()||event.button!==0)return;event.preventDefault();const startX=event.clientX,startWidth=workspacePanel.getBoundingClientRect().width,resizer=$('#workspace-resizer');
  const guide=el('div','workspace-resize-guide');guide.style.left=startX+'px';document.body.appendChild(guide);document.body.classList.add('workspace-resizing');resizer.classList.add('workspace-resizer--active');
  let nextWidth=startWidth,frameRequest=0;
  const paint=()=>{frameRequest=0;guide.style.transform='translate3d('+(startWidth-nextWidth)+'px,0,0)';};
  const move=moveEvent=>{nextWidth=clampWorkspaceWidth(startWidth+startX-moveEvent.clientX);if(!frameRequest)frameRequest=requestAnimationFrame(paint);};
  const stop=()=>{if(frameRequest)cancelAnimationFrame(frameRequest);applyWorkspaceWidth(nextWidth,true);guide.remove();document.body.classList.remove('workspace-resizing');resizer.classList.remove('workspace-resizer--active');window.removeEventListener('pointermove',move);window.removeEventListener('pointerup',stop);window.removeEventListener('pointercancel',stop);};
  window.addEventListener('pointermove',move);window.addEventListener('pointerup',stop);window.addEventListener('pointercancel',stop);
});
workspaceResizer.addEventListener('keydown',event=>{if(mobileLayout())return;let width=workspacePanel.getBoundingClientRect().width;const step=event.shiftKey?50:20;if(event.key==='ArrowLeft')width+=step;else if(event.key==='ArrowRight')width-=step;else if(event.key==='Home')width=workspaceWidthBounds().min;else if(event.key==='End')width=workspaceWidthBounds().max;else return;event.preventDefault();applyWorkspaceWidth(width,true);});
window.addEventListener('resize',()=>{if(!mobileLayout())applyWorkspaceWidth(workspacePanel.getBoundingClientRect().width||WORKSPACE_WIDTH_DEFAULT);});

// ── cumulative stats ──
// Session-scoped counters accumulated while this page is open. They reset on
// session rotation (/new, /resume) so the stats card labeled "session cost"
// never quietly turns into an all-sessions running total.
let cumulativeTokens=0, cumulativeCacheHit=0, cumulativeCacheMiss=0;
let sessionCostQuote=null;
function resetCumulativeStats(){cumulativeTokens=0;cumulativeCacheHit=0;cumulativeCacheMiss=0;sessionCostQuote=null;}
function openStats(){
  if(!prepareOrdinaryOverlay())return;
  const modal=$('#stats-modal');
  if(!modal)return;
  // populate from latest status
  fetch('/status').then(r=>r.json()).then(s=>{
    $('#stats-model').textContent=s.label||'-';
    $('#stats-sessions').textContent=sessionCount||'0';
    const total=cumulativeTokens||s.used||0;
    $('#stats-total-tokens').textContent=fmtTok(total);
    const hit=cumulativeCacheHit||s.cacheHit||0, miss=cumulativeCacheMiss||s.cacheMiss||0;
    const rate=hit+miss>0?Math.round(hit/(hit+miss)*100)+'%':'0%';
    $('#stats-cache-rate').textContent=rate;
    const quote=s.sessionCostQuote;
    if(quote?.displayStatus==='bucketed'||quote?.aggregateMode==='currency_buckets')$('#stats-total-cost').textContent=__('multi_currency');
    else if(quote?.selected?.amount)$('#stats-total-cost').textContent=fmtMoney(Number(quote.selected.amount),quote.selected.currency);
    else $('#stats-total-cost').textContent='—';
    $('#stats-balance').textContent=s.balance?.display||'-';
    if(s.window){const pct=Math.min(100,Math.round(s.used/s.window*100));$('#stats-ctx-fill').style.width=pct+'%';$('#stats-ctx-used').textContent=fmtTok(s.used)+' tokens';$('#stats-ctx-window').textContent=fmtTok(s.window)+' tokens';}
  }).catch(()=>{});
  modal.style.display='flex';
}
$('#stats-modal-close').onclick=()=>{const m=$('#stats-modal');if(m)m.style.display='none';};
$('#stats-modal').onclick=e=>{if(e.target===e.currentTarget){const m=$('#stats-modal');if(m)m.style.display='none';}};

// ── session delete ──
document.addEventListener('click',e=>{
  const del=e.target.closest('.session-del');
  if(!del)return;
  e.stopPropagation();
  const name=del.dataset.name;
  const target=sessionsCache.find(s=>s.name===name);
  if(target&&target.current){showNotice(__('cannot_delete_active'),'warn');return;}
  openDeleteSession(name);
});
$('#delete-modal-close').onclick=()=>closeDeleteSession();
$('#delete-cancel').onclick=()=>closeDeleteSession();
$('#delete-modal').onclick=e=>{if(e.target===e.currentTarget)closeDeleteSession();};
$('#delete-confirm').onclick=()=>{
  const name=pendingDeleteSession;
  if(!name)return;
  closeDeleteSession();
  post('/delete-session',{name}).then(async r=>{
    if(!r.ok){showNotice((await r.text()).trim()||('HTTP '+r.status),'warn');}
    loadSessions();
  }).catch(()=>showNotice(__('delete_failed'),'warn'));
};

// welcome examples
$$('.welcome__ex').forEach(btn=>{btn.onclick=()=>{input.value=btn.dataset.prompt;send();};});

// ── theme toggle (light/dark; persisted per-browser, dark default) ──
const themeBtn=$('#btn-theme');
function applyTheme(mode){
  if(mode==='light'){document.documentElement.setAttribute('data-theme','light');localStorage.setItem('baize-theme','light');themeBtn.title=__('theme_switch_dark');}
  else{document.documentElement.removeAttribute('data-theme');localStorage.setItem('baize-theme','dark');themeBtn.title=__('theme_switch_light');}
  themeBtn.setAttribute('aria-label',themeBtn.title);
}
applyTheme(document.documentElement.getAttribute('data-theme')==='light'?'light':'dark');
themeBtn.onclick=()=>{const next=document.documentElement.getAttribute('data-theme')==='light'?'dark':'light';setStorageValue('baize-theme-preference',next);applyTheme(next);};

// ── settings drawer ──
const settingsDrawer=$('#settings-drawer'),settingsBackdrop=$('#settings-backdrop'),settingsButton=$('#btn-settings'),settingsForm=$('#settings-form');
let settingsRevision='',settingsSnapshot=null,settingsDirty=false;
function storageValue(key,fallback){try{return localStorage.getItem(key)||fallback;}catch{return fallback;}}
function setStorageValue(key,value){try{localStorage.setItem(key,value);}catch{}}
function effectiveThemePreference(){
  const pref=storageValue('baize-theme-preference',storageValue('baize-theme','dark'));
  if(pref==='auto')return window.matchMedia&&window.matchMedia('(prefers-color-scheme: light)').matches?'light':'dark';
  return pref==='light'?'light':'dark';
}
function applyAppearanceSettings(){
  applyTheme(effectiveThemePreference());
  document.documentElement.dataset.density=storageValue('baize-density','comfortable');
}
function openSettings(){
  if(!ordinaryOverlayAllowed())return;
  closeWorkspacePanel({restoreFocus:false});closeOrdinaryModals();
  settingsDrawer.classList.add('settings-drawer--open');settingsDrawer.setAttribute('aria-hidden','false');settingsBackdrop.hidden=false;settingsButton.setAttribute('aria-expanded','true');
  if(!settingsDirty)loadSettings();
}
function closeSettings({preserveDraft=false,restoreFocus=false}={}){const wasOpen=settingsDrawer?.classList.contains('settings-drawer--open');settingsDrawer?.classList.remove('settings-drawer--open');settingsDrawer?.setAttribute('aria-hidden','true');if(settingsBackdrop)settingsBackdrop.hidden=true;settingsButton?.setAttribute('aria-expanded','false');if(!preserveDraft)settingsDirty=false;if(wasOpen&&restoreFocus)settingsButton?.focus();}
function settingsState(text,tone){const state=$('#settings-state');state.textContent=text||'';state.dataset.tone=tone||'';}
function option(select,value,label){const o=document.createElement('option');o.value=value;o.textContent=label;select.appendChild(o);}
function fillModelSetting(select,models,value,allowEmpty){
  select.innerHTML='';if(allowEmpty)option(select,'',__('auto'));
  (models||[]).forEach(model=>option(select,model,model));
  if(value&&!Array.from(select.options).some(o=>o.value===value))option(select,value,value);
  select.value=value||'';
}
function populateSettings(view){
  settingsSnapshot=view;settingsRevision=view.revision||'';const value=view.global||{};const models=value.models||[];
  fillModelSetting($('#setting-default-model'),models,value.defaultModel,false);fillModelSetting($('#setting-planner-model'),models,value.plannerModel,true);fillModelSetting($('#setting-subagent-model'),models,value.subagentModel,true);
  Object.entries(value).forEach(([name,val])=>{const field=settingsForm.elements.namedItem(name);if(field&&field.tagName!=='SELECT'&&field.tagName!=='INPUT')return;if(field&&name!=='defaultModel'&&name!=='plannerModel'&&name!=='subagentModel')field.value=String(val);});
  $('#setting-theme').value=storageValue('baize-theme-preference',storageValue('baize-theme','dark'));
  $('#setting-density').value=storageValue('baize-density','comfortable');
  $('#setting-reasoning-display').value=storageValue('baize-reasoning-display','closed');
  $('#setting-subagent-preview').value=storageValue('baize-subagent-preview','full');
  $('#setting-subagent-collapse').checked=storageValue('baize-subagent-auto-collapse','true')!=='false';
  const overridden=Array.isArray(view.overridden)?view.overridden:[];
  let message='';let tone='';
  if(view.applyError){message=view.applyError;tone='danger';}
  else if(view.apply==='pending'){message=__('settings_pending');tone='warn';}
  else if(view.apply==='applying'){message=__('settings_applying');}
  else if(overridden.length){const effective=view.effective||{};message=__('settings_overridden').replace('{fields}',overridden.map(name=>name+' → '+String(effective[name]??'')).join(', '));tone='warn';}
  settingsState(message,tone);$('#settings-retry').hidden=!view.applyError;settingsDirty=false;
}
async function loadSettings(){
  try{const response=await fetch('/settings');if(!response.ok)throw new Error((await response.text()).trim()||('HTTP '+response.status));populateSettings(await response.json());}
  catch(error){settingsState(error instanceof Error?error.message:String(error),'danger');}
}
function runtimeSettingsPayload(){
  const data=new FormData(settingsForm);const payload={revision:settingsRevision};
  ['defaultModel','plannerModel','subagentModel','subagentEffort','defaultApprovalMode','reasoningLanguage'].forEach(name=>payload[name]=String(data.get(name)||''));
  ['maxSubagentDepth','maxSubagentConcurrency','maxParallelWriters'].forEach(name=>payload[name]=Number(data.get(name)));
  payload.compactRatio=Number(data.get('compactRatio'));return payload;
}
function saveAppearanceSettings(){
  setStorageValue('baize-theme-preference',$('#setting-theme').value);setStorageValue('baize-density',$('#setting-density').value);setStorageValue('baize-reasoning-display',$('#setting-reasoning-display').value);setStorageValue('baize-subagent-preview',$('#setting-subagent-preview').value);setStorageValue('baize-subagent-auto-collapse',String($('#setting-subagent-collapse').checked));applyAppearanceSettings();
}
['setting-theme','setting-density','setting-reasoning-display','setting-subagent-preview','setting-subagent-collapse'].forEach(id=>{const field=$('#'+id);if(field)field.addEventListener('change',saveAppearanceSettings);});
settingsForm.addEventListener('input',event=>{if(event.target?.name)settingsDirty=true;});
settingsForm.addEventListener('change',event=>{if(event.target?.name)settingsDirty=true;});
settingsForm.onsubmit=async event=>{
  event.preventDefault();const payload=runtimeSettingsPayload();
  if(payload.defaultApprovalMode==='yolo'&&settingsSnapshot?.global?.defaultApprovalMode!=='yolo'&&!window.confirm(__('settings_yolo_confirm')))return;
  saveAppearanceSettings();settingsState(__('settings_applying'));
  try{
    const response=await fetch('/settings',{method:'PATCH',headers:{'content-type':'application/json'},body:JSON.stringify(payload)});
    if(response.status===409){populateSettings(await response.json());settingsState(__('settings_conflict'),'warn');return;}
    if(!response.ok)throw new Error((await response.text()).trim()||('HTTP '+response.status));
    const view=await response.json();populateSettings(view);if(view.apply==='applied')settingsState(__('settings_applied'));else if(view.apply==='pending')settingsState(__('settings_pending'),'warn');
  }catch(error){settingsState(error instanceof Error?error.message:String(error),'danger');}
};
$('#settings-retry').onclick=async()=>{try{const response=await post('/settings/apply',{});if(!response.ok&&response.status!==202)throw new Error((await response.text()).trim()||('HTTP '+response.status));await loadSettings();}catch(error){settingsState(error instanceof Error?error.message:String(error),'danger');}};
settingsButton.onclick=openSettings;$('#settings-close').onclick=()=>closeSettings({restoreFocus:true});settingsBackdrop.onclick=()=>closeSettings({restoreFocus:true});
if(window.matchMedia){window.matchMedia('(prefers-color-scheme: light)').addEventListener?.('change',()=>{if(storageValue('baize-theme-preference','dark')==='auto')applyAppearanceSettings();});}
applyAppearanceSettings();

// One Escape closes only the highest-priority ordinary surface. Approval and
// ask shelves own Escape while a decision is pending; the rewind picker owns
// its keys in its earlier capture-phase handler.
document.addEventListener('keydown',event=>{
  if(event.key!=='Escape'||event.defaultPrevented||decisionInteractionLocked||typeof calSelected!=='undefined'&&calSelected)return;
  const consume=()=>{event.preventDefault();event.stopPropagation();};
  if(imageViewer){consume();closeImageViewer();return;}
  if(settingsDrawer.classList.contains('settings-drawer--open')){consume();closeSettings({restoreFocus:true});return;}
  for(const [id,close] of [['delete-modal',closeDeleteSession],['models-modal',closeModels],['branches-modal',closeBranches],['stats-modal',()=>{const modal=$('#stats-modal');if(modal)modal.style.display='none';}]]){const modal=$('#'+id);if(modal&&modal.style.display!=='none'){consume();close();return;}}
  if(workspaceOpen){consume();closeWorkspacePanel({restoreFocus:true});return;}
  if(sidebar.classList.contains('sidebar--open')){consume();closeSidebar();}
},true);

// initial fetch
fetchStatus();
refreshCheckpointAvailability();
updateActionAvailability();
