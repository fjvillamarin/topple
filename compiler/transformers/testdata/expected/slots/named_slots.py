class Layout(BaseView):
    def __init__(self, *, children=None, header=None, footer=None):
        super().__init__()
        self.children = children
        self.header = header
        self.footer = footer

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("header", render_child(self.header) if self.header is not None else ""))
        _div_children_2000.append(el("main", render_child(self.children) if self.children is not None else ""))
        _div_children_2000.append(el("footer", render_child(self.footer) if self.footer is not None else ""))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

