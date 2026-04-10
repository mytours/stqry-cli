class ProjectsResource:
    def __init__(self, http):
        self._http = http

    def list(self, **query) -> list:
        resp = self._http.get("/api/public/projects", params=query or None)
        return resp.get("projects", [])

    def get(self, id: str) -> dict:
        resp = self._http.get(f"/api/public/projects/{id}", params=None)
        return resp.get("project", resp)
