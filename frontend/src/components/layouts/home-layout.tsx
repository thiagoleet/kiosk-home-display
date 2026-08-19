import type { ReactNode } from "react";

type HomeLayoutProps = {
  children: ReactNode;
};

export function HomeLayout({ children }: HomeLayoutProps) {
  return <section data-layout="home">{children}</section>;
}
