class ConditionalView(BaseView):
    def __init__(self, show_message, message):
        super().__init__()
        self.show_message = show_message
        self.message = message

    def _render(self) -> Element:
        if self.show_message:
            el("div", escape(self.message))
        else:
            el("div", escape("No message"))

