class MediaResource:
    def __init__(self, http):
        self._http = http

    def list(self, **query) -> list:
        resp = self._http.get("/api/public/media_items", params=query or None)
        return resp.get("media_items", [])

    def get(self, id: str) -> dict:
        resp = self._http.get(f"/api/public/media_items/{id}", params=None)
        return resp.get("media_item", resp)

    def create(self, **fields) -> dict:
        resp = self._http.post("/api/public/media_items", {"media_item": fields})
        return resp.get("media_item", resp)

    def update(self, id: str, **fields) -> dict:
        resp = self._http.patch(f"/api/public/media_items/{id}", {"media_item": fields})
        return resp.get("media_item", resp)

    def delete(self, id: str, **query) -> None:
        self._http.delete(f"/api/public/media_items/{id}", params=query or None)
