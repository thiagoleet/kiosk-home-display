import { HomeView } from "../views/home-view";

import { ActivityWidget } from "../widgets/activity-widget";
import { PrinterWidget } from "../widgets/printer-widget";

export function HomeLayout() {
  return (
    <section className="home-layout">
      <div className="home-layout__main">
        <HomeView />
      </div>

      <footer className="home-layout__footer">
        <PrinterWidget />
        <ActivityWidget />
      </footer>
    </section>
  );
}
