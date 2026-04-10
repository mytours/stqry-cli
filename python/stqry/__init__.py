import os
from stqry._http import HttpClient
from stqry.collections import CollectionsResource
from stqry.screens import ScreensResource
from stqry.media import MediaResource
from stqry.codes import CodesResource
from stqry.projects import ProjectsResource


class Client:
    def __init__(self, api_url: str | None = None, token: str | None = None):
        api_url = api_url or os.environ.get("STQRY_API_URL")
        token = token or os.environ.get("STQRY_API_TOKEN")
        if not api_url or not token:
            raise ValueError(
                "api_url and token are required. Pass them explicitly or set "
                "STQRY_API_URL and STQRY_API_TOKEN environment variables."
            )
        self._http = HttpClient(api_url, token)
        self.collections = CollectionsResource(self._http)
        self.screens = ScreensResource(self._http)
        self.media = MediaResource(self._http)
        self.codes = CodesResource(self._http)
        self.projects = ProjectsResource(self._http)


__all__ = ["Client"]
