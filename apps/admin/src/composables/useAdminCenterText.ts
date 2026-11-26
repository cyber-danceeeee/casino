import { useI18n } from "vue-i18n";

export function useAdminCenterText(center: "userCenter" | "fundsCenter") {
  const { locale, tm } = useI18n();

  return (text: string) => {
    if (locale.value !== "zh-CN") return text;
    const messages = tm(`${center}.text`) as Record<string, string>;
    return messages[text] || text;
  };
}
