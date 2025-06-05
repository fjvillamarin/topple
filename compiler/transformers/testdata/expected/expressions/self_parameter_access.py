class UserProfile(BaseView):
    def __init__(self, username, email):
        super().__init__()
        self.username = username
        self.email = email

    def _render(self) -> Element:
        return el("div", [el("h1", ""), el("p", "")])

