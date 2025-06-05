class ItemList(BaseView):
    def __init__(self, items):
        super().__init__()
        self.items = items

    def _render(self) -> Element:
        return el("ul", "")

