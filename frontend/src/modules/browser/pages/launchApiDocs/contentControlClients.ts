export const DOC_CONTROL_CLI = `# antctl 命令行

\`antctl\` 适合批量脚本、定时任务和本机终端。它通过 Launch API 操作 Ant Browser，不直接读写数据库。使用前先启动桌面应用，并确认 \`GET /api/health\` 可用。

## 连接与认证

默认连接 \`http://127.0.0.1:19876\`。如果应用开启 API Key，将密钥放在环境变量中，不要写入脚本或提交到 Git。

\`\`\`bash
export ANT_BROWSER_URL="http://127.0.0.1:19876"
export ANT_BROWSER_API_KEY="YOUR_API_KEY"

./build/antctl health
./build/antctl profiles list
./build/antctl proxies list
\`\`\`

未开启 API Key 时，不设置 \`ANT_BROWSER_API_KEY\` 即可。也可用 \`--base-url\` 和 \`--api-key\`，但命令行参数可能出现在进程列表中，优先使用环境变量。

## 实例生命周期

\`\`\`bash
# 新建实例并绑定已有代理节点
./build/antctl profiles create --name slot-01 --proxy-id PROXY_ID

# 复制实例，原实例保留；输出中读取新 profileId
./build/antctl profiles copy PROFILE_ID --name slot-01-next --mode auto_fingerprint

# 更新代理、启动、停止
./build/antctl profiles update PROFILE_ID --proxy-id PROXY_ID
./build/antctl profiles start PROFILE_ID
./build/antctl profiles stop PROFILE_ID
\`\`\`

\`profiles copy\` 默认使用 \`auto_fingerprint\`，创建独立用户数据目录并返回新实例 ID。它继承源实例的代理绑定；如需不同节点，在启动前执行 \`profiles update\`。旧实例不会被删除。

## 批量创建与自动化

\`\`\`bash
./build/antctl profiles create-batch profiles.json
./build/antctl automation scripts list
./build/antctl automation scripts get SCRIPT_ID
./build/antctl automation scripts run SCRIPT_ID --selector '{"profileId":"PROFILE_ID"}'
./build/antctl automation runs list --limit 20
\`\`\`

\`profiles.json\` 是 \`ProfileInput\` 对象数组，例如 \`[{"profileName":"slot-01","proxyId":"PROXY_ID"}]\`。批量创建可能部分成功，失败时也要检查输出中的已创建项，避免重复创建。

## 安全边界

- \`profiles delete PROFILE_ID --confirm\` 会删除已停止实例；保留历史实例时不要执行。
- 多个实例共用代理端口时，切换该端口的出口 IP 可能影响它们后续运行。
- CLI 输出会隐藏用户目录、代理凭证等敏感字段；需要完整配置时使用受控的本地 API。
`

export const DOC_CONTROL_MCP = `# ant-mcp Agent 接入

\`ant-mcp\` 是 stdio MCP 服务，适合 Claude、Antigravity 和其他支持 MCP 的 Agent。它与 \`antctl\` 使用同一套操作目录，最终都调用本地 Launch API。

## 客户端配置

\`\`\`json
{
  "ant-browser": {
    "command": "/ABSOLUTE_PATH/Ant-Browser/build/ant-mcp",
    "args": [],
    "env": {
      "ANT_BROWSER_URL": "http://127.0.0.1:19876",
      "ANT_BROWSER_API_KEY": ""
    }
  }
}
\`\`\`

将 \`command\` 改为当前机器上的绝对路径。没开启 API Key 时可保持空字符串；开启后在客户端的安全凭证设置中提供。新增工具会通过 MCP 的 \`tools/list\` 自动发现，原有配置通常不需要改。

## 可用工具

| 操作 | MCP 工具 |
|------|----------|
| 健康检查、代理列表 | \`health\`、\`list_proxy_nodes\` |
| 实例查询 | \`list_instances\`、\`get_instance\` |
| 实例创建与修改 | \`create_instance\`、\`update_instance\`、\`create_instances_batch\` |
| 实例复制 | \`copy_instance\` |
| 实例启动、停止、删除 | \`start_instance\`、\`stop_instance\`、\`delete_instance\` |
| 自动化脚本 | \`list_automation_scripts\`、\`get_automation_script\`、\`run_automation_script\` |
| 执行记录 | \`list_automation_runs\` |

## 推荐 Agent 流程

1. 先调用 \`health\`，再用 \`list_proxy_nodes\` 与 \`list_instances\` 验证 ID。
2. 要保留旧实例时调用 \`copy_instance\`，读取返回的新 \`profileId\`，不要根据名称猜测。
3. 如需切换节点，使用 \`update_instance\` 指定新实例的 \`proxy_id\`。
4. 调用 \`start_instance\`，确认可用后再运行脚本。
5. 用 \`stop_instance\` 停止旧实例；不要自动删除历史数据。

当请求涉及周期轮换或并发槽位时，把调度逻辑放在外部脚本/Agent；Ant Browser 负责实例生命周期与自动化执行。
`

export const DOC_CONTROL_SKILL = `# ant-browser-control Skill

仓库内的 \`skills/ant-browser-control/\` 是给 Agent 阅读的操作指南，不是 Ant Browser 的运行依赖。它总结 API、CLI、MCP 的能力及安全使用顺序。

## 在本机启用

\`\`\`bash
mkdir -p ~/.codex/skills
ln -s /ABSOLUTE_PATH/Ant-Browser/skills/ant-browser-control ~/.codex/skills/ant-browser-control
\`\`\`

如果目标路径已经存在，先检查它是否指向同一个仓库，不要直接覆盖。其他 Agent 客户端可按其 Skill 目录规范复制或软链接该文件夹。

## 什么时候使用

- 让 Agent 查询或修改实例、代理绑定、自动化脚本。
- 开发外部轮换调度器，需要确认复制实例、返回新 ID 与保留旧数据的约束。
- 排查 CLI/MCP 与 Launch API 的功能映射。

Skill 的详细能力表在 \`skills/ant-browser-control/references/api-surface.md\`。实际控制路径始终是 \`Agent → ant-mcp / antctl → Launch API → Ant Browser\`。
`
