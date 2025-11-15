from topple.psx import BaseView, Element, el, escape, fragment
class StaticAttributes(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("img", "", {"src": "/logo.png", "alt": "Company Logo"}))
        _div_children_2000.append(el("input", "", {"type": "text", "placeholder": "Enter your name", "readonly": True}))
        _div_children_2000.append(el("button", "Submit", {"type": "submit", "disabled": True}))
        _root_children_1000.append(el("div", _div_children_2000, {"class": "container", "id": "main-content", "data-testid": "app"}))
        return fragment(_root_children_1000)

