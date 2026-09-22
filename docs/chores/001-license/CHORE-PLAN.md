# CHORE-PLAN — 001-license

**Вертикаль:** chore (no-bump). **Файлы:** `LICENSE` (новый, BSD-3-Clause — как у `pinout-asyncapi`).
**Команда верификации:** `test -f LICENSE && head -1 LICENSE | grep -q BSD`.
**Откат:** `git revert`. Правка нужна для консистентности экосистемы (в README уже заявлена открытая лицензия).
