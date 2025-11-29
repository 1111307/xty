from fastapi import FastAPI, HTTPException
from typing import Dict, Any

#uvicorn backend:app --host 127.0.0.1 --port 8008 --reload
# --- 玉简数据库（知识库）数据 ---
# 在实际项目中，这里会替换成连接 SQLite 或 PostgreSQL 的代码
JADE_DATABASE: Dict[str, Any] = {
    "101": {"title": "长生诀", "level": "天阶", "content": "长生诀，乃是夺天地造化之奇书...练此功者，生生不息。"},
    "102": {"title": "烈火掌", "level": "黄阶", "content": "气沉丹田，引火入掌，掌出如龙，焚尽八荒。适合火灵根修炼。"},
    "103": {"title": "御剑术基础", "level": "玄阶", "content": "以气御剑，人剑合一。初学者需先感悟剑意，方可尝试离手御剑。"},
}

app = FastAPI(title="玉简知识库 API")

# 接口1：搜索玉简（工具 Tool 对应接口）
@app.get("/api/search")
async def search_slips(query: str):
    """根据关键词搜索玉简"""
    results = []
    for id, data in JADE_DATABASE.items():
        # 简单地检查关键词是否在标题或内容中
        if query in data["title"] or query in data["content"]:
            results.append({"id": id, "title": data["title"], "level": data["level"]})
    return {"results": results}

# 接口2：读取具体玉简内容（Resource 对应接口）
@app.get("/api/slip/{slip_id}")
async def get_slip_content(slip_id: str):
    """根据 ID 读取玉简详细内容"""
    if slip_id in JADE_DATABASE:
        return JADE_DATABASE[slip_id]
    # 如果 ID 不存在，抛出 404 异常
    raise HTTPException(status_code=404, detail="玉简未找到或已损毁")