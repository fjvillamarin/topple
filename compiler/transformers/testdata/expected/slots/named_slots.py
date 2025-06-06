class Layout(BaseView):
    def __init__(self, *, children=None, header=None, footer=None):
        super().__init__()
        self.children = children
        self.header = header
        self.footer = footer

    def _render(self) -> Element:
        return el("div", [el("header", render_child(self.header) if self.header is not None else ""), el("main", render_child(self.children) if self.children is not None else ""), el("footer", render_child(self.footer) if self.footer is not None else "")])

