# Generowanie ODS - Pipeline

> **Cel:** Pełny pipeline tworzenia pliku `Pokemon-GO-{account}.ods` gotowego do pracy w LibreOffice.
> Proces jest idempotentny - może być uruchomiony wielokrotnie, nadpisując plik.

---

## Pipeline Overview

```plantuml
@startuml

|Actor|
start
:Podaj nazwę konta X;

|System|
:Load Gamemaster\nParse JSON;
:Przygotuj wpisy Pokemon\n1598 gatunków z Gamemaster\nFilter: released + force_released;
:Normalizuj nazwy\n(Forma) => ' - Forma'\nDla VLOOKUP kompatybilności;
:Buduj arkusze ODS:\n1. Kolekcja [1000 wierszy]\n2. Pokedex [1598 wierszy]\n3. Pokemon [1598 wierszy]\n4. Ruchy [347 wierszy]\n5. Jak używać\n6. Listy [walidacje]\n7. Pokedex [checkboxes];
:Buduj walidacje:\n- Listy dla Data Validation\n- Checkbox ranges\n- VLOOKUP formulas;
:Generuj content.xml\nStyles, tables, formulas;
:Pakuj do ZIP ODS\nmimetype + content.xml + styles.xml;

|Actor|
:Pokemon-GO-{X}.ods gotowy;
stop

@enduml
```

---

## Szczegóły Kroków

### 1. Load Gamemaster

```plantuml
@startuml

start
:Plik gamemaster.json\nistnieje lokalnie?;

if (Tak) then (tak)
else (nie)
    :Pobierz z GitHub\nPvPoke Gamemaster URL;
    :Zapisz data/gamemaster.json;
endif

:Parsuj JSON;

if (Struktura OK?) then (tak)
    :Gamemaster gotowy;
    stop
else (nie)
    :BŁĄD: zła struktura;
    stop
endif

@enduml
```

### 2. Prepare Pokemon Entries

```plantuml
@startuml

start
:Dla każdego pokemon w gamemaster.pokemon;

repeat
    if (released == true\nLUB force_released?) then (tak)
        :Wyciągnij:\n- dex, speciesId, speciesName\n- types, baseStats\n- fastMoves[], chargedMoves[]\n- tags[];
        :Zamień moveId => name\nvia moves map;
        :Normalizuj nazwę\n(Forma) => ' - Forma';
        :Zbuduj entry:\n{dex, speciesId, speciesName,\ntypes, baseStats, fastNames,\nchargedNames, fastList, chargedList, tags};
        :Dodaj do listy entries;
    else (nie)
    endif
repeat while (Next)

:Sortuj:\n1. dex [int], 2. speciesName;
:Entries gotowe: 1598+;
stop

@enduml
```

### 3. Build Sheets - Arkusz Kolekcja (Główny)

```plantuml
@startuml

|System|
start
:Build Kolekcja Sheet;

:Zdefiniuj kolumny A-AA:\nUID, Gatunek, Własna, Typ1/2, CP,\nIV Atk/Def/HP, IV%, Fast, Ch1, Ch2,\nShiny/Lucky/Shadow/Purified/Dynamax/Traded,\nInne, Data, Przeznaczenie, Priorytet,\nDecyzja, Cel, Powód, Uwagi;

:Wiersz nagłówka\nStyl: bold, center, wrap;

:1000 wierszy danych:\n- Formuły VLOOKUP na Gatunek => Pokedex\n- Typ1/2 z Pokedex\n- IV% = (Atk+Def+HP)/45*100\n- Data Validation na checkboxy\n- Freeze pane: wiersz 1;

:Listy walidacji (Arkusz Listy):\n- Przeznaczenie: PvP, Rajdy, Gym, Trade, Transfer, Kolekcja\n- Decyzja: ROZWIŃ, PRETENDENT, TRADE, PROFESOR, ZACHOWAJ\n- Tagi: Shiny, Lucky, Shadow, Purified, Dynamax, Traded\n- Priorities: 1-5;

:Formuły VLOOKUP:\n- Typ1 = VLOOKUP(Gatunek, Pokedex, col_Typ1, 0)\n- Typ2 = VLOOKUP(Gatunek, Pokedex, col_Typ2, 0)\n- IFERROR => '[brak]';

:Kolekcja Sheet gotowy;
stop

@enduml
```

### 4. Build Sheets - Arkusz Pokedex (1598 gatunków)

```plantuml
@startuml

start
:Build Pokedex Sheet;

:Kolumny:\nDex, Gatunek, Forma, Typ1, Typ2,\nCaught, Lucky, Shiny, Hundo, Shadow,\nPurified, Dynamax, Mega, Tagi, Notatki;

:Dla każdego entry z entries[];

:AutoSync z Collection:\n- Caught = X jeśli species w Kolekcja\n- Lucky = X jeśli Lucky w Kolekcja\n- Shiny = X jeśli Shiny w Kolekcja\n- Hundo = X jeśli IV=100% w Kolekcja\n- Shadow = X jeśli Shadow w Kolekcja\n- Purified = X jeśli Purified w Kolekcja\n- Dynamax = X jeśli Dynamax w Kolekcja\n- Mega = X jeśli Mega w Kolekcja;

:Checkbox ranges:\nData Validation => Listy.\nCheckbox range references;

:Freeze: wiersz 1, kolumna B\nGatunek zawsze widoczny;

:Pokedex Sheet gotowy: 1598 wierszy;
stop

@enduml
```

### 5. Build Sheets - Pokemon, Ruchy, Listy

```plantuml
@startuml

|Pokemon|
start
:Kolumny: Dex, Gatunek, Typ1/2,\nAtak, Obrona, HP, MaxCP L40/L50,\nSzybkie ruchy, Ładowane ruchy, Tagi;
:Dla każdego entry;
:1598 wierszy;

|Ruchy|
:Kolumny: Ruch, Typ, Rodzaj, Moc,\nEnergia koszt/zysk, Cooldown, Tury,\nDPE, DPT, EPT;
:Dla każdego move z gamemaster.moves;
:347 wierszy;

|Jak używać|
:Instrukcja użytkownika\nMarkdown w komórkach;

|Listy|
:Przeznaczenie list;
:Decyzja list;
:Tagi list;
:Priorytet list;
:Checkbox ranges\nreferencje do komórek Pokedex;

|System|
:Wszystkie arkusze gotowe;
stop

@enduml
```

### 6. Generate content.xml + Package ODS

```plantuml
@startuml

start
:Generate ODS XML;

:Generuj styles.xml\n- Cell styles: header, data, number, pct\n- Number formats: percent, integer, decimal\n- Alignment, borders, wrap;

:Generuj content.xml\n- office:spreadsheet\n- table:table per sheet\n- table:table-column (widths)\n- table:table-row + table:table-cell\n- office:value-type + office:value\n- Formuły: table:formula\n- Data Validation: table:content-validation-name;

:Generuj settings.xml\n- View settings\n- Freeze panes config;

:Generuj manifest.xml\n- File entries + media types;

:Zapisz mimetype\napplication/vnd.oasis.opendocument.spreadsheet;

:Pakuj do ZIP:\nmimetype [bez kompresji]\ncontent.xml\nstyles.xml\nsettings.xml\nmanifest.xml;

:Zapisz Pokemon-GO-{account}.ods;
:ODS gotowy;
stop

@enduml
```

---

## Normalizacja Nazw - Kluczowa Logika

```plantuml
@startuml

start
:Input: 'Arcanine [Hisuian]';

:Regex: \\s*\\(.*\\) => ' - \\1';
:Trim: double space => single space;
:Output: 'Arcanine - Hisuian';

note right
  Dlaczego?
  - Usunięcie kwadratowych nawiasów
    uniemożliwia VLOOKUP
  - Spójność z Pokedex entry
  - Czytelność dla użytkownika
end note

stop

@enduml
```

---

## Walidacja Wygenerowanego ODS

```plantuml
@startuml

start
:Verify Generated ODS;

:Plik istnieje?;
:Rozmiar > 10KB?;
:Otwórz ZIP;

if (content.xml OK?) then (tak)
else (nie)
    :FAIL;
    stop
endif

:Parsuj XML;
:Arkusze = 7+?;
:Wiersze Kolekcja >= 100?;
:Komórki >= 1000?;
:VLOOKUP count >= 2xdata_rows?;
:Walidacje istnieją?;
:SUCCESS + stats;
:ODS Valid;
stop

@enduml
```

---

## Parametry Wejściowe / Wyjściowe

| Parametr | Typ | Opis |
|----------|-----|------|
| `account` | string | Nazwa konta (np. `Adzik`, `Shizaya09`) |
| `gamemaster_path` | path | Opcjonalny - override ścieżki gamemaster.json |
| `output_path` | path | Opcjonalny - override ścieżki wyjściowej |
| `force` | bool | Wymuś regenerację nawet jeśli plik istnieje |

---

## Zależności Zewnętrzne

| Źródło | Format | Częstotliwość aktualizacji |
|--------|--------|---------------------------|
| PvPoke Gamemaster | JSON | Co wydanie Game Mastera (co ~2 tygodnie) |
| PvPoke Rankings | JSON (cached) | Co miesiąc / --force |

---

## Błędy i Obsługa

| Błąd | Reakcja |
|------|---------|
| Brak Gamemaster JSON | Pobierz z GitHub |
| Brak `released` flag | Użyj `force_released` set |
| Błąd XML generation | Log + FAIL |
| Zła struktura ODS | Log + FAIL (weryfikacja wykryje) |
