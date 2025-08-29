import { atom } from "nanostores";
import type { Theme } from "./theme-provider";

export const themeStore = atom<Theme>("light");
