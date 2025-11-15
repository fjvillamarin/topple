class NestedElements(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("h1", escape("Title")))
        _div_children_2000.append(el("p", escape("Content")))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

