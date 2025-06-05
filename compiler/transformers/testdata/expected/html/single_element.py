class SingleElement(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("p", escape("Simple paragraph"))

