from runtime import BaseView, el, escape, Element, FragmentElement, fragment, render_child
class MultiRoot(BaseView):
    def __init__(self, title: str, content: str):
        super().__init__()
        self.title = title
        self.content = content

    def _render(self) -> Element:
        _view_children_1000 = []
        _view_children_1000.append(el("h1", escape(self.title)))
        _view_children_1000.append(el("p", escape(self.content)))
        _view_children_1000.append(el("div", el("span", "Additional content")))
        return fragment(_view_children_1000)

