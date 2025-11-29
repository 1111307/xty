from mcp.server.fastmcp import FastMCP
import requests # <-- 新增: 引入同步 requests 库
import asyncio
import httpx # 保持导入，但仅用于关闭全局客户端

# --- 配置 ---
# 确保你的 FastAPI 知识库后端正在 8008 端口运行，并且你已将其绑定到 127.0.0.1
# 注意：BACKEND_URL 现在是基础地址，不包含 API 路径
BACKEND_URL = "http://127.0.0.1:8008"
# -------------

# **注意：由于使用 requests，我们不再需要 GLOBAL_HTTP_CLIENT，但为了 cleanup 保持其定义**
GLOBAL_HTTP_CLIENT = httpx.AsyncClient(
    timeout=5.0,
    verify=False
)


# 创建 MCP 服务器实例，名称为 "demo-fmcp"
mcp = FastMCP("demo-fmcp", json_response=True)


# ==================================
# 1. 本地工具 (测试工具调用是否成功)
# ==================================
@mcp.tool()
def add(a: int, b: int) -> int:
    """计算两个数字的和"""
    return a + b


# ==================================
# 2. 玉简知识库工具 (核心功能)
# ==================================
@mcp.tool()
async def search_jade_library(keyword: str) -> str:
    """
    【必须调用】在藏经阁中搜索功法、武技、秘籍等玉简内容。
    当用户询问关于**功法、武技、心法、秘籍、属性（如火、水、金）、门派**等信息时，
    请使用此工具。
    
    Args:
        keyword: 搜索关键词，如 "剑", "火", "长生诀"
    """
    try:
        # **关键修正：使用 asyncio.to_thread 调用同步的 requests.get**
        # requests.get 是同步阻塞操作，to_thread() 将其放入线程池中执行
        response = await asyncio.to_thread(
            requests.get, 
            f"{BACKEND_URL}/api/search", 
            params={"query": keyword}, 
            timeout=5.0
        )
        
        response.raise_for_status() # 检查 HTTP 状态码，如果 >= 400 则抛出异常
        data = response.json()
        
        # 格式化结果给大模型阅读
        slips = data.get("results", [])
        if not slips:
            return "神识探查中，藏经阁未找到与此关键词相关的玉简。"
        
        result_text = "找到以下玉简 (请使用 jade://ID 协议读取详情):\n"
        for slip in slips:
            result_text += f"- ID: {slip['id']} | 名称: {slip['title']} | 等级: {slip['level']}\n"
        return result_text
            
    except Exception as e:
        # 捕获连接失败或 4xx/5xx 错误
        # 这里的 e 可能是 requests.exceptions.RequestException 或 requests.exceptions.HTTPError
        return f"神识探查失败，可能与藏经阁连接中断 (Error: {type(e).__name__} - {str(e)})"

# ==================================
# 3. 玉简知识库资源 (核心功能)
# ==================================
@mcp.resource("jade://{slip_id}")
async def read_jade_slip(slip_id: str) -> str:
    """
    读取指定 ID 的玉简详细内容。
    """
    # **关键修正：使用 asyncio.to_thread 调用同步的 requests.get**
    try:
        # 代理请求到后端 /api/slip/{id}
        response = await asyncio.to_thread(
            requests.get, 
            f"{BACKEND_URL}/api/slip/{slip_id}",
            timeout=5.0
        )
        
        if response.status_code == 404:
            return f"ID 为 {slip_id} 的玉简不存在或已被封印。"
        
        response.raise_for_status()
        data = response.json()
        
        # 返回详细内容
        return f"""
=== 玉简详情 ===
名称：{data.get('title', 'N/A')}
等级：{data.get('level', 'N/A')}
---
{data.get('content', '内容缺失')}
=== 结束 ===
        """
    except Exception as e:
        return f"读取玉简失败: {str(e)}"


if __name__ == "__main__":
    # 确保 FastAPI 后端已在 127.0.0.1:8008 运行
    try:
        mcp.run()
    finally:
        # 保持关闭 GLOBAL_HTTP_CLIENT 的逻辑，虽然 requests 不依赖它，但保持代码完整性
        asyncio.run(GLOBAL_HTTP_CLIENT.aclose())