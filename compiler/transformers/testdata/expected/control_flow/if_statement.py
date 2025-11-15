class ConditionalView(BaseView):
    def __init__(self, show_message, message):
        super().__init__()
        self.show_message = show_message
        self.message = message

    def _render(self) -> Element:
        _root_children_1000 = []
        if self.show_message:
            _root_children_1000.append(el("div", escape(self.message)))
        else:
            _root_children_1000.append(el("div", escape("No message")))
        return fragment(_root_children_1000)

