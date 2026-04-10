class CollectionsResource:
    def __init__(self, http):
        self._http = http

    def list(self, **query) -> list:
        resp = self._http.get("/api/public/collections", params=query or None)
        return resp.get("collections", [])

    def get(self, id: str) -> dict:
        resp = self._http.get(f"/api/public/collections/{id}", params=None)
        return resp.get("collection", resp)

    def create(self, **fields) -> dict:
        resp = self._http.post("/api/public/collections", {"collection": fields})
        return resp.get("collection", resp)

    def update(self, id: str, **fields) -> dict:
        resp = self._http.patch(f"/api/public/collections/{id}", {"collection": fields})
        return resp.get("collection", resp)

    def delete(self, id: str) -> None:
        self._http.delete(f"/api/public/collections/{id}", params=None)

    def list_items(self, collection_id: str, **query) -> list:
        path = f"/api/public/collections/{collection_id}/collection_items"
        resp = self._http.get(path, params=query)
        return resp.get("collection_items", [])

    def get_item(self, collection_id: str, item_id: str) -> dict:
        path = f"/api/public/collections/{collection_id}/collection_items/{item_id}"
        resp = self._http.get(path, params=None)
        return resp.get("collection_item", resp)

    def create_item(self, collection_id: str, **fields) -> dict:
        path = f"/api/public/collections/{collection_id}/collection_items"
        resp = self._http.post(path, {"collection_item": fields})
        return resp.get("collection_item", resp)

    def update_item(self, collection_id: str, item_id: str, **fields) -> dict:
        path = f"/api/public/collections/{collection_id}/collection_items/{item_id}"
        resp = self._http.patch(path, {"collection_item": fields})
        return resp.get("collection_item", resp)

    def delete_item(self, collection_id: str, item_id: str) -> None:
        path = f"/api/public/collections/{collection_id}/collection_items/{item_id}"
        self._http.delete(path, params=None)

    def reorder_items(self, collection_id: str, item_ids: list) -> None:
        path = f"/api/public/collections/{collection_id}/collection_items/update_positions"
        self._http.post(path, {"ids": item_ids})
