from topple.psx import BaseView, Element, el, escape, fragment
class Greeting(BaseView):
    def __init__(self, name: str, age: int=25):
        super().__init__()
        self.name = name
        self.age = age

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("h1", f"Hello, {escape(self.name)}!"))
        _div_children_2000.append(el("p", f"You are {escape(self.age)} years old."))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

