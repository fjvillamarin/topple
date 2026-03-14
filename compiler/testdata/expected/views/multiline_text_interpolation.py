from topple.psx import BaseView, Element, el, escape, fragment, raw
class Greeting(BaseView):
    def __init__(self, name: str, title: str):
        super().__init__()
        self.name = name
        self.title = title

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("p", f"Hello {escape(self.name)}, welcome to {escape(self.title)}!"))
        _div_children_2000.append(el("span", f"Your name is {escape(self.name)} and your title is {escape(self.title)}."))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

