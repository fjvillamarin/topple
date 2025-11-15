from topple.psx import BaseView, Element, el, escape, fragment, raw
class HTMXBasic(BaseView):
    def __init__(self, user_id: int):
        super().__init__()
        self.user_id = user_id

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("h1", "HTMX Demo"))
        _div_children_2000.append(el("button", escape("Load Data"), {"hx-get": "/api/data", "hx-target": "#content", "hx-swap": "innerHTML"}))
        _div_children_2000.append(el("div", escape("Loading..."), {"id": "content", "hx-get": "/api/initial", "hx-trigger": "load"}))
        _form_children_3000 = []
        _form_children_3000.append(el("input", "", {"type": "hidden", "name": "user_id", "value": escape(f"{self.user_id}")}))
        _form_children_3000.append(el("input", "", {"type": "text", "name": "message", "placeholder": "Enter message..."}))
        _form_children_3000.append(el("button", "Send", {"type": "submit"}))
        _div_children_2000.append(el("form", _form_children_3000, {"hx-post": "/api/submit", "hx-target": "#result", "hx-swap": "outerHTML"}))
        _div_children_2000.append(el("div", "", {"id": "result"}))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

