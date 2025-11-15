class UserCard(BaseView):
    def __init__(self, user: User, show_email: bool=False):
        super().__init__()
        self.user = user
        self.show_email = show_email

    def _render(self) -> Element:
        return el("div", el("h2", escape(self.user.name)), {"class": "user-card"})

