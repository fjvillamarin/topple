class EmptyView(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return fragment([])

