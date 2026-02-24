from topple.psx import BaseView, Element, el, escape, fragment, raw
from ..components.cards import FeatureCard
from ..utils.data import FEATURES
class FeaturesGrid(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        for feature in FEATURES:
            _div_children_2000.append(FeatureCard(icon=feature["icon"], title=feature["title"], description=feature["desc"]))
        _root_children_1000.append(el("div", _div_children_2000, {"class": "grid md:grid-cols-3 gap-6 mb-20"}))
        return fragment(_root_children_1000)

