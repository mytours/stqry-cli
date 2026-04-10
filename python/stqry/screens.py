class ScreensResource:
    def __init__(self, http):
        self._http = http

    def list(self, **query) -> list:
        resp = self._http.get("/api/public/screens", params=query or None)
        return resp.get("screens", [])

    def get(self, id: str) -> dict:
        resp = self._http.get(f"/api/public/screens/{id}", params=None)
        return resp.get("screen", resp)

    def create(self, **fields) -> dict:
        resp = self._http.post("/api/public/screens", {"screen": fields})
        return resp.get("screen", resp)

    def update(self, id: str, **fields) -> dict:
        resp = self._http.patch(f"/api/public/screens/{id}", {"screen": fields})
        return resp.get("screen", resp)

    def delete(self, id: str) -> None:
        self._http.delete(f"/api/public/screens/{id}", params=None)

    # Story sections

    def list_sections(self, screen_id: str, **query) -> list:
        path = f"/api/public/screens/{screen_id}/story_sections"
        resp = self._http.get(path, params=query or None)
        return resp.get("story_sections", [])

    def get_section(self, screen_id: str, section_id: str) -> dict:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}"
        resp = self._http.get(path, params=None)
        return resp.get("story_section", resp)

    def create_section(self, screen_id: str, **fields) -> dict:
        path = f"/api/public/screens/{screen_id}/story_sections"
        resp = self._http.post(path, {"story_section": fields})
        return resp.get("story_section", resp)

    def update_section(self, screen_id: str, section_id: str, **fields) -> dict:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}"
        resp = self._http.patch(path, {"story_section": fields})
        return resp.get("story_section", resp)

    def delete_section(self, screen_id: str, section_id: str) -> None:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}"
        self._http.delete(path, params=None)

    def reorder_sections(self, screen_id: str, section_ids: list) -> None:
        path = f"/api/public/screens/{screen_id}/story_sections/update_positions"
        self._http.post(path, {"section_ids": section_ids})

    # Generic section sub-items

    def list_sub_items(self, screen_id: str, section_id: str, sub_item_type: str) -> list:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}/{sub_item_type}"
        resp = self._http.get(path, params=None)
        return resp.get(sub_item_type, [])

    def create_sub_item(self, screen_id: str, section_id: str, sub_item_type: str, singular_key: str, **fields) -> dict:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}/{sub_item_type}"
        resp = self._http.post(path, {singular_key: fields})
        return resp.get(singular_key, resp)

    def update_sub_item(self, screen_id: str, section_id: str, sub_item_type: str, item_id: str, singular_key: str, **fields) -> dict:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}/{sub_item_type}/{item_id}"
        resp = self._http.patch(path, {singular_key: fields})
        return resp.get(singular_key, resp)

    def delete_sub_item(self, screen_id: str, section_id: str, sub_item_type: str, item_id: str) -> None:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}/{sub_item_type}/{item_id}"
        self._http.delete(path, params=None)
