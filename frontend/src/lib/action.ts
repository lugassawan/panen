import type { ActionType } from "./types";

export const ACTION_LABELS: Record<ActionType, string> = {
  BUY: "Buy",
  AVERAGE_DOWN: "Average Down",
  AVERAGE_UP: "Average Up",
  SELL_EXIT: "Sell (Exit)",
  SELL_STOP: "Sell (Stop Loss)",
  HOLD: "Hold",
};

export const ACTION_DESCRIPTIONS: Record<ActionType, string> = {
  BUY: "Open a new position in this stock",
  AVERAGE_DOWN: "Buy more at a lower price to reduce average cost",
  AVERAGE_UP: "Add to position for dividend income",
  SELL_EXIT: "Take profit and close position",
  SELL_STOP: "Cut loss and close position",
  HOLD: "Maintain current position",
};
