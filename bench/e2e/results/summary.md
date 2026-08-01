# dizz e2e benchmark summary

## Overall medians
| metric | with_dizz | without_dizz | delta% |
|---|---|---|---|
| input tokens | 37,877 | 35,626 | +6.3% |
| output tokens | 4,742 | 3,780 | +25.4% |
| cache reads | 153,344 | 125,312 | +22.4% |
| duration (s) | 53.0 | 45.0 | +17.8% |
| tool calls | 17.0 | 16.0 | +6.2% |
| files changed | 2.0 | 2.0 | +0.0% |

## Success rate
| task | with_dizz | without_dizz |
|---|---|---|
| 01_deadcode | 5/5 | 5/5 |
| 02_bugfix | 5/5 | 5/5 |
| 03_todos | 5/5 | 5/5 |
| 04_plan | 5/5 | 4/5 |
| 05_refactor | 4/5 | 3/5 |

## Per task - input tokens
| task | with_dizz | without_dizz | delta | delta% |
|---|---|---|---|---|
| 01_deadcode | 38,041 | 36,200 | 1,841 | +5.1% |
| 02_bugfix | 13,722 | 13,807 | -85.0 | -0.6% |
| 03_todos | 16,340 | 16,599 | -259.0 | -1.6% |
| 04_plan | 43,848 | 37,365 | 6,483 | +17.4% |
| 05_refactor | 91,510 | 83,305 | 8,205 | +9.8% |

## Per task - duration (s)
| task | with_dizz | without_dizz | delta | delta% |
|---|---|---|---|---|
| 01_deadcode | 40.0 | 37.0 | 3.0 | +8.1% |
| 02_bugfix | 13.0 | 14.0 | -1.0 | -7.1% |
| 03_todos | 45.0 | 49.0 | -4.0 | -8.2% |
| 04_plan | 71.0 | 57.0 | 14.0 | +24.6% |
| 05_refactor | 72.0 | 149.0 | -77.0 | -51.7% |

## Per task - tool calls
| task | with_dizz | without_dizz | delta | delta% |
|---|---|---|---|---|
| 01_deadcode | 17.0 | 17.0 | 0.0 | +0.0% |
| 02_bugfix | 5.0 | 6.0 | -1.0 | -16.7% |
| 03_todos | 13.0 | 14.0 | -1.0 | -7.1% |
| 04_plan | 23.0 | 19.0 | 4.0 | +21.1% |
| 05_refactor | 23.0 | 21.0 | 2.0 | +9.5% |

## Honesty gate
- with_dizz success rate: 96.0%
- without_dizz success rate: 88.0%
- OK: dizz success rate is within 20 points of the control.

# dizz e2e sequence benchmark

One project per (condition, run); tasks run as consecutive opencode sessions on the same project. Session 1 is the first run; later sessions are subsequent runs on accumulated state.

## Per session - input tokens
| session | with_dizz | without_dizz | delta | delta% |
|---|---|---|---|---|
| 1 | 41,567 | 36,790 | 4,777 | +13.0% |
| 2 | 16,163 | 13,611 | 2,552 | +18.7% |
| 3 | 23,088 | 16,868 | 6,220 | +36.9% |

## Per session - output tokens
| session | with_dizz | without_dizz | delta | delta% |
|---|---|---|---|---|
| 1 | 5,713 | 3,281 | 2,432 | +74.1% |
| 2 | 907.0 | 545.0 | 362.0 | +66.4% |
| 3 | 7,272 | 4,206 | 3,066 | +72.9% |

## Per session - cache reads
| session | with_dizz | without_dizz | delta | delta% |
|---|---|---|---|---|
| 1 | 141,184 | 107,008 | 34,176 | +31.9% |
| 2 | 76,928 | 52,480 | 24,448 | +46.6% |
| 3 | 256,256 | 155,008 | 101,248 | +65.3% |

## Per session - duration (s)
| session | with_dizz | without_dizz | delta | delta% |
|---|---|---|---|---|
| 1 | 56.0 | 35.0 | 21.0 | +60.0% |
| 2 | 19.0 | 14.0 | 5.0 | +35.7% |
| 3 | 67.0 | 41.0 | 26.0 | +63.4% |

## Cumulative input tokens after each session
| after session | with_dizz | without_dizz | delta% |
|---|---|---|---|
| 1 | 41,567 | 36,790 | +13.0% |
| 2 | 57,730 | 51,289 | +12.6% |
| 3 | 79,475 | 67,931 | +17.0% |

## First run vs subsequent runs
| metric | with_dizz | without_dizz | delta% |
|---|---|---|---|
| input_tokens - first run | 41,567 | 36,790 | +13.0% |
| input_tokens - avg per subsequent session | 19,813 | 15,730 | +26.0% |
| input_tokens - total | 79,475 | 67,931 | +17.0% |
| output_tokens - first run | 5,713 | 3,281 | +74.1% |
| output_tokens - avg per subsequent session | 4,361 | 2,363 | +84.6% |
| output_tokens - total | 13,101 | 8,355 | +56.8% |
| cache_read - first run | 141,184 | 107,008 | +31.9% |
| cache_read - avg per subsequent session | 174,784 | 103,744 | +68.5% |
| cache_read - total | 506,368 | 342,784 | +47.7% |
| duration_s - first run | 56.0 | 35.0 | +60.0% |
| duration_s - avg per subsequent session | 45.0 | 27.5 | +63.6% |
| duration_s - total | 136.0 | 97.0 | +40.2% |

## Success rate by session
| session | with_dizz | without_dizz |
|---|---|---|
| 1 | 5/5 | 5/5 |
| 2 | 5/5 | 5/5 |
| 3 | 5/5 | 5/5 |

## Honesty gate (full-cell completion)
- with_dizz full-cell success rate: 100.0% (5/5)
- without_dizz full-cell success rate: 100.0% (5/5)
- OK: dizz success rate is within 20 points of the control.
