from topple.psx import BaseView, Element, el, escape, fragment
class UserList(BaseView):
    def __init__(self, users: list):
        super().__init__()
        self.users = users

    def _render(self) -> Element:
        _root_children_2000 = []
        _ul_children_3000 = []
        for user in self.users:
            _ul_children_3000.append(el("li", escape(user)))
        _root_children_2000.append(el("ul", _ul_children_3000))
        return fragment(_root_children_2000)

