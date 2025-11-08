from topple.psx import BaseView, Element, el, escape, fragment
class HTMXBasic(BaseView):
    def __init__(self, user_id: int):
        super().__init__()
        self.user_id = user_id

    def _render(self) -> Element:
        return el("div", [el("h1", "HTMX Demo"), el("button", ["", ""], {"hx-get": "/api/data", "hx-target": "#content", "hx-swap": "innerHTML"}), el("div", ["", ""], {"id": "content", "hx-get": "/api/initial", "hx-trigger": "load"}), el("form", [el("input", "", {"type": "hidden", "name": "user_id", "value": escape(f"{self.user_id}")}), el("input", "", {"type": "text", "name": "message", "placeholder": "Enter message..."}), el("button", "Send", {"type": "submit"})], {"hx-post": "/api/submit", "hx-target": "#result", "hx-swap": "outerHTML"}), el("div", "", {"id": "result"})])

