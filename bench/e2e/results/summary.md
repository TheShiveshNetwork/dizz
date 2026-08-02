# dizz e2e benchmark summary

## Overall medians
| metric | with_dizz | without_dizz | delta% |
|---|---|---|---|
| input tokens | 37,877 | 35,626 | +6.3% |
| output tokens | 4,742 | 3,780 | +25.4% |
| cache reads | 153,344 | 125,312 | +22.4% |
| duration (s) | 51.0 | 38.7 | +31.6% |
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
| 01_deadcode | 37.6 | 32.6 | 5.1 | +15.5% |
| 02_bugfix | 9.8 | 10.3 | -0.6 | -5.4% |
| 03_todos | 42.9 | 47.0 | -4.1 | -8.7% |
| 04_plan | 67.3 | 48.6 | 18.7 | +38.5% |
| 05_refactor | 65.6 | 58.3 | 7.3 | +12.5% |

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

## dizz usage (with_dizz)
| task | runs w/ dizz | runs w/ dizz context | avg dizz calls |
|---|---|---|---|
| 01_deadcode | 0/5 | 0/5 | 0.0 |
| 02_bugfix | 1/5 | 1/5 | 0.2 |
| 03_todos | 1/5 | 1/5 | 0.2 |
| 04_plan | 2/5 | 2/5 | 0.6 |
| 05_refactor | 3/5 | 3/5 | 1.8 |

# dizz e2e sequence benchmark

One project per (condition, run); tasks run as consecutive opencode sessions on the same project. Session 1 is the first run; later sessions are subsequent runs on accumulated state.

## Per session - input tokens
| session | with_dizz | without_dizz | delta | delta% |
|---|---|---|---|---|
| 1 | 40,476 | 31,855 | 8,621 | +27.1% |
| 2 | 8,040 | 7,485 | 555.0 | +7.4% |
| 3 | 7,390 | 5,214 | 2,176 | +41.7% |

## Per session - output tokens
| session | with_dizz | without_dizz | delta | delta% |
|---|---|---|---|---|
| 1 | 4,631 | 2,927 | 1,704 | +58.2% |
| 2 | 594.0 | 533.0 | 61.0 | +11.4% |
| 3 | 6,798 | 4,682 | 2,116 | +45.2% |

## Per session - cache reads
| session | with_dizz | without_dizz | delta | delta% |
|---|---|---|---|---|
| 1 | 140,672 | 106,752 | 33,920 | +31.8% |
| 2 | 61,696 | 45,056 | 16,640 | +36.9% |
| 3 | 261,376 | 187,264 | 74,112 | +39.6% |

## Per session - duration (s)
| session | with_dizz | without_dizz | delta | delta% |
|---|---|---|---|---|
| 1 | 53.0 | 32.0 | 21.0 | +65.6% |
| 2 | 14.0 | 13.0 | 1.0 | +7.7% |
| 3 | 71.0 | 47.0 | 24.0 | +51.1% |

## Cumulative input tokens after each session
| after session | with_dizz | without_dizz | delta% |
|---|---|---|---|
| 1 | 40,476 | 31,855 | +27.1% |
| 2 | 47,773 | 39,321 | +21.5% |
| 3 | 55,603 | 43,988 | +26.4% |

## First run vs subsequent runs
| metric | with_dizz | without_dizz | delta% |
|---|---|---|---|
| input_tokens - first run | 40,476 | 31,855 | +27.1% |
| input_tokens - avg per subsequent session | 7,715 | 6,373 | +21.1% |
| input_tokens - total | 55,603 | 43,988 | +26.4% |
| output_tokens - first run | 4,631 | 2,927 | +58.2% |
| output_tokens - avg per subsequent session | 3,758 | 2,608 | +44.1% |
| output_tokens - total | 12,673 | 8,106 | +56.3% |
| cache_read - first run | 140,672 | 106,752 | +31.8% |
| cache_read - avg per subsequent session | 161,536 | 120,128 | +34.5% |
| cache_read - total | 463,744 | 334,464 | +38.7% |
| duration_s - first run | 53.0 | 32.0 | +65.6% |
| duration_s - avg per subsequent session | 42.0 | 32.0 | +31.2% |
| duration_s - total | 141.0 | 101.0 | +39.6% |

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

## dizz usage (with_dizz sessions)
| session | runs w/ dizz | runs w/ dizz context | avg dizz calls | commands |
|---|---|---|---|---|
| 1 | 3/5 | 3/5 | 0.8 | dizz context, dizz context 2>&1 | head -100, dizz context 2>&1 | head -50, dizz log 2>&1 | head -120 |
| 2 | 1/5 | 1/5 | 0.2 | dizz context 2>&1 | head -50 |
| 3 | 5/5 | 5/5 | 1.4 | dizz config show --guardrails, dizz context, dizz context 2>&1 || echo "dizz not available", dizz context 2>&1 | head -50, dizz intent list 2>&1 || echo "dizz unavailable" |
- Session 1 init via `dizz context`: 3/5 runs (skill.md compliance)
