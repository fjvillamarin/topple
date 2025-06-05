class NestedElements(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", [el("h1", escape("Title")), el("p", escape("Content"))])

