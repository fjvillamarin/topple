from topple.psx import BaseView, Element, el, escape, fragment
class MultiRoot(BaseView):
    def __init__(self, title: str, content: str):
        super().__init__()
        self.title = title
        self.content = content

    def _render(self) -> Element:
        _root_children_1000 = []
        _root_children_1000.append(el("h1", escape(self.title)))
        _root_children_1000.append(el("p", escape(self.content)))
        _root_children_1000.append(el("div", el("span", "Additional content")))
        return fragment(_root_children_1000)

