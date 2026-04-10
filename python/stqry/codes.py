class CodesResource:
    def __init__(self, http):
        self._http = http

    def list(self, **query) -> list:
        resp = self._http.get("/api/public/codes", params=query or None)
        return resp.get("codes", [])

    def get(self, id: str) -> dict:
        resp = self._http.get(f"/api/public/codes/{id}", params=None)
        return resp.get("code", resp)

    def create(self, **fields) -> dict:
        resp = self._http.post("/api/public/codes", {"code": fields})
        return resp.get("code", resp)

    def update(self, id: str, **fields) -> dict:
        resp = self._http.patch(f"/api/public/codes/{id}", {"code": fields})
        return resp.get("code", resp)

    def delete(self, id: str) -> None:
        self._http.delete(f"/api/public/codes/{id}", params=None)
