class Greeting(BaseView):
    def __init__(self, name: str, title: str = "Mr."):
        super().__init__()
        self.name = name
        self.title = title

    def _render(self) -> Element:
        return el("h1", escape(f"Hello, {title} {name}"))

