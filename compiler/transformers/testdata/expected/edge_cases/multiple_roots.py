class MultipleRoots(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_1000 = []
        _root_children_1000.append(el("h1", escape("First")))
        _root_children_1000.append(el("p", escape("Second")))
        _root_children_1000.append(el("div", escape("Third")))
        return fragment(_root_children_1000)

