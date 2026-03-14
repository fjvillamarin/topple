from topple.psx import BaseView, Element, el, escape, fragment, raw
class Article(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_1000 = []
        _article_children_2000 = []
        _article_children_2000.append(el("p", f"This is a longer paragraph that spans multiple lines."))
        _article_children_2000.append(el("div", f"Some introductory text here followed by more content on separate lines.", {"class": "content"}))
        _root_children_1000.append(el("article", _article_children_2000))
        return fragment(_root_children_1000)

class SimpleMultiline(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("p", f"Hello world from PSX!")

