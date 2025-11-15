from topple.psx import BaseView, Element, el, escape, fragment, raw
from ..components import PageHeader, CodeComparison
from ..sections import FeaturesGrid, ExamplesSection
class LandingPage(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_1000 = []
        _html_children_2000 = []
        _head_children_3000 = []
        _head_children_3000.append(el("meta", "", {"charset": "UTF-8"}))
        _head_children_3000.append(el("meta", "", {"name": "viewport", "content": "width=device-width, initial-scale=1.0"}))
        _head_children_3000.append(el("title", "Topple - PSX (Multi-File Demo)"))
        _head_children_3000.append(el("script", "", {"src": "https://cdn.tailwindcss.com"}))
        _html_children_2000.append(el("head", _head_children_3000))
        _body_children_4000 = []
        _div_children_5000 = []
        _div_children_5000.append(PageHeader(title="Topple", subtitle="PSX: HTML-like syntax in Python", description="Write components using familiar JSX syntax, render on the server using Python", cta_text="Get Started", cta_url="https://github.com/topple-lang/topple"))
        _div_children_5000.append(CodeComparison())
        _div_children_5000.append(FeaturesGrid())
        _div_children_5000.append(ExamplesSection())
        _body_children_4000.append(el("div", _div_children_5000, {"class": "container mx-auto px-4 max-w-6xl"}))
        _html_children_2000.append(el("body", _body_children_4000, {"class": "bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 min-h-screen py-16"}))
        _root_children_1000.append(el("html", _html_children_2000, {"lang": "en"}))
        return fragment(_root_children_1000)

