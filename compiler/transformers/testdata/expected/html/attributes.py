class ElementWithAttributes(BaseView):
    def __init__(self, css_class: str):
        super().__init__()
        self.css_class = css_class

    def _render(self) -> Element:
        return el("div", escape("Content with attributes"), {"class": escape(self.css_class), "id": "main-div"})

