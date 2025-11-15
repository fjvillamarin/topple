class ItemList(BaseView):
    def __init__(self, items):
        super().__init__()
        self.items = items

    def _render(self) -> Element:
        _root_children_1000 = []
        _ul_children_2000 = []
        for item in self.items:
            _ul_children_2000.append(el("li", escape(item)))
        _root_children_1000.append(el("ul", _ul_children_2000))
        return fragment(_root_children_1000)

