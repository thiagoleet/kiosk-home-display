import { useFeature } from "@/hooks/use-feature";
import { usePrinter } from "@/hooks/use-printer";
import { useTranslation } from "@/hooks/use-translation";
import { ThemeIcon } from "../theme/theme-icon";

export function PrinterWidget() {
  const { state, currentJob } = usePrinter();
  const isPrinting = state === "printing";
  const { t } = useTranslation();

  const printerEnabled = useFeature("printer");

  // On a box with no printer the widget would sit on "ready" for good,
  // describing a device that is not there.
  if (!printerEnabled) {
    return null;
  }

  return (
    <section
      className="widget printer-widget"
      data-state={state}
    >
      <div className="widget-title">
        <ThemeIcon name="printer" />
        <span>{t("printer.title")}</span>
      </div>

      <div className="printer-widget__state">
        {isPrinting ? (
          <>
            <ThemeIcon
              name="loading"
              className="spin"
            />

            <span>{t("printer.printing")}</span>
          </>
        ) : (
          <>
            <ThemeIcon name="notification.success" />

            <span>{t("printer.ready")}</span>
          </>
        )}
      </div>

      <p className="printer-widget__job">
        {currentJob?.name ?? t("printer.noActiveJobs")}
      </p>
    </section>
  );
}
