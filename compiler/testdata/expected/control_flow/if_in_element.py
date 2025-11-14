from topple.psx import BaseView, Element, el, escape, fragment
class Greeting(BaseView):
    def __init__(self, is_admin: bool, name: str):
        super().__init__()
        self.is_admin = is_admin
        self.name = name

    def _render(self) -> Element:
        _root_children_2000 = []
        _div_children_3000 = []
        if self.is_admin:
            _div_children_3000.append(el("h1", f"Admin: {escape(self.name)}"))
        else:
            _div_children_3000.append(el("p", f"User: {escape(self.name)}"))
        _root_children_2000.append(el("div", _div_children_3000))
        return fragment(_root_children_2000)

