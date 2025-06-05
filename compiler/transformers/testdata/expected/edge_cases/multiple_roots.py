class MultipleRoots(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        el("h1", escape("First"))
        el("p", escape("Second"))
        el("div", escape("Third"))

