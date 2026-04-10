import requests


class HttpClient:
    def __init__(self, api_url: str, token: str):
        self._base_url = api_url.rstrip("/")
        self._session = requests.Session()
        self._session.headers.update({
            "X-Api-Token": token,
            "Content-Type": "application/json",
            "Accept": "application/json",
        })

    def get(self, path: str, params: dict = None) -> dict:
        resp = self._session.get(self._base_url + path, params=params, timeout=30)
        resp.raise_for_status()
        return resp.json()

    def post(self, path: str, json: dict = None) -> dict:
        resp = self._session.post(self._base_url + path, json=json, timeout=30)
        resp.raise_for_status()
        return resp.json() if resp.content else {}

    def patch(self, path: str, json: dict) -> dict:
        resp = self._session.patch(self._base_url + path, json=json, timeout=30)
        resp.raise_for_status()
        return resp.json()

    def delete(self, path: str, params: dict = None) -> None:
        resp = self._session.delete(self._base_url + path, params=params, timeout=30)
        resp.raise_for_status()
