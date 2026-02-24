from topple.psx import BaseView, Element, el, escape, fragment, raw
class Greeting(BaseView):
    def __init__(self, name: str):
        super().__init__()
        self.name = name

    def _render(self) -> Element:
        return el("h1", f"Hello, {escape(self.name)}!")

class Profile(BaseView):
    def __init__(self, name: str, title: str):
        super().__init__()
        self.name = name
        self.title = title

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("h2", escape(self.name)))
        _div_children_2000.append(el("p", escape(self.title)))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

class Dashboard(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_3000 = []
        name = "Dashboard"
        title = "Admin Panel"
        _div_children_4000 = []
        _div_children_4000.append(el("h3", escape(name)))
        _div_children_4000.append(el("span", escape(title)))
        _root_children_3000.append(el("div", _div_children_4000))
        return fragment(_root_children_3000)

