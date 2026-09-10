# Backup i Weryfikacja ODS

> **Cel:** Strategia backupu, weryfikacji integralności i odzyskiwania danych.
> Zasada: **Ostatni backup zawsze musi być grywalny** - to Twoja polisa ubezpieczeniowa.

---

## Strategia Backupu

```plantuml
@startuml

|Triggers|
start
:Ręczne: uruchom komendę Backup\nLUB Przed zmianą w ODS\nLUB Przed regeneracją ODS;

|System|
:Create Backup;
:Weryfikuj backup:\n- ZIP otwieralny\n- content.xml exists + parseable\n- Rozmiar > 100B;

if (Weryfikacja?) then (FAIL)
    :Usuń backup + FAIL;
    stop
else (OK)
endif

:Usuń > 5 najstarszych\nSort by mtime DESC\nKeep [0:5], Delete [5:];

|Actor|
:Backup gotowy;
stop

@enduml
```

---

## Proces Tworzenia Backupu

```plantuml
@startuml

start
:Backup ODS;
:Rozwiąż ścieżkę:\nPokemon-GO-{account}.ods;

if (Plik istnieje?) then (tak)
else (nie)
    :FAIL: Brak pliku źródłowego;
    stop
endif

:Timestamp: YYYYMMDD_HHMMSS;
:backups/Pokemon-GO-{X}.ods.bak.{timestamp};
:Kopiuj plik\n[zachowuje metadata];

:Weryfikuj backup:\nOtwórz jako ZIP;

if (content.xml exists?) then (tak)
else (nie)
    :Delete backup + FAIL;
    stop
endif

if (Rozmiar > 100B?) then (tak)
else (nie)
    :Delete backup + FAIL;
    stop
endif

:Prune Old Backups\nSort by mtime DESC\nKeep [0:5], Delete [5:];
:SUCCESS: {backup_path};
stop

@enduml
```

---

## Weryfikacja ODS

```plantuml
@startuml

start
:Weryfikuj ODS;
:Otwórz plik;

if (ZIP valid?) then (tak)
else (nie)
    :FAIL: Nieprawidłowy ZIP;
    stop
endif

:Odczytaj content.xml;
:Parsuj XML;
:Policz arkusze [table:table];
:Policz wiersze [table:table-row];
:Policz komórki [table:table-cell];

if (Wiersze >= 100?) then (tak)
else (nie)
    :FAIL: Za mało wierszy;
    stop
endif

if (Komórki >= 1000?) then (tak)
else (nie)
    :FAIL: Za mało komórek;
    stop
endif

if (Nazwy arkuszy poprawne?) then (tak)
else (nie)
    :FAIL: Brak arkuszy;
    stop
endif

:Log: SUCCESS + stats\n- plik, rozmiar, arkusze, wiersze, komórki;
stop

@enduml
```

---

## Przywracanie z Backupu

```plantuml
@startuml

start
:Przywróć ODS;
:Arg: ścieżka backupu [docelowa ścieżka ODS];

:Rozwiąż backup:\n- Pełna ścieżka LUB\n- Nazwa pliku [szukaj w backups/];

if (Znaleźiono?) then (tak)
else (nie)
    :FAIL: Backup nie znaleziony;
    stop
endif

:Weryfikuj backup:\nZIP + content.xml + size;

if (Weryfikacja OK?) then (tak)
else (nie)
    :FAIL: Backup uszkodzony;
    stop
endif

:Skopiuj jako Pokemon-GO-{X}.ods\n[nadpisuje bieżący];

:Zweryfikuj przywrócony plik\n[logika weryfikacji];

if (Zweryfikuj OK?) then (tak)
else (nie)
    :FAIL: Przywrócony uszkodzony;
    stop
endif

:SUCCESS/FAIL + ścieżka;
stop

@enduml
```

---

## Polityka Retencji

```plantuml
@startuml

left to right direction

rectangle "Backup Directory" {
    [B1\n20260901_100000] as B1
    [B2\n20260902_143000] as B2
    [B3\n20260903_091500] as B3
    [B4\n20260904_164500] as B4
    [B5\n20260905_112000] as B5
    [B6\n20260906_080000] as B6
    [DELETE] as Trash
}

B6 --> Trash : DELETE

note bottom of B1
  SORT by mtime DESC
  Keep [0:5] = B6,B5,B4,B3,B2
  Delete [5:] = B1
end note

@enduml
```

| Parametr | Wartość |
|----------|---------|
| Max kopii | 5 per konto |
| Sortowanie | Według mtime (modification time) DESC |
| Usuwane | Wszystkie poza 5 najnowszymi |
| Nazewnictwo | `Pokemon-GO-{account}.ods.bak.YYYYMMDD_HHMMSS` |

---

## Workflow Bezpieczeństwa (Kiedy robić backup?)

```plantuml
@startuml

|Actor|
start
:Edytuj ODS w LibreOffice\nDodaj/usuń pokemony;

|System|
if (Backup przed edycją?) then (tak)
    :Uruchom komendę Backup;
else (nie)
    :RYZYKO: Brak backupu!;
endif

|Actor|
:Regeneruj ODS;

|System|
if (Backup przed regeneracją?) then (tak)
    :Uruchom komendę Backup;
else (nie)
    :RYZYKO: Brak backupu!;
endif

|Actor|
:Analizuj;

|System|
:Bezpieczne: Analiza tylko czyta;

|Actor|
:Aktualizuj Meta;

|System|
:Bezpieczne: Tylko meta_rankings.json;

stop

@enduml
```

> **ZASADA ZŁOTA:** **Backup PRZED KAŻDĄ ZMIANĄ W ODS**. Analiza i aktualizacja meta są read-only - nie wymagają backupu.

---

## Scenariusze Awaryjne

| Scenariusz | Rozwiązanie |
|------------|-----------|
| ODS uszkodzony po edycji w LO | Przywróć najnowszy backup |
| Regeneracja ODS z błędnymi danymi | Przywróć backup sprzed regeneracji |
| Przypadkowe usunięcie ODS | Przywróć najnowszy backup |
| Błąd w regeneracji | Backup nie dotknięty → przywróć |
| Chcesz cofnąć do wersji z X dni temu | Przywróć konkretny backup z datą |

---

## Weryfikacja Ciągłości (Po przywróceniu)

```plantuml
@startuml

start
:Przywrócony ODS;
:Uruchom weryfikację;

if (Weryfikacja OK?) then (tak)
else (nie)
    :FAIL: Nie przywracaj dalej;
    stop
endif

:Otwórz w LibreOffice;
:Sprawdź wizualnie:\n- Liczba pokemonów\n- Kluczowe UID\n- Walidacje działają;
:Porównaj kluczowe UID\nz collection JSON snapshot;
:CONTINUITY CONFIRMED;
stop

@enduml
```

---

## Metryki Backupu

| Metryka | Wartość docelowa |
|---------|-----------------|
| Czas backupu | < 2 sekundy |
| Czas weryfikacji | < 1 sekunda |
| Czas przywracania | < 3 sekundy |
| Maksymalny rozmiar backupu | ~500 KB (zależny od ODS) |
| Prędkość prune | < 0.1 sekundy |

---

## Testowanie Backupu (Chaos Engineering)

```plantuml
@startuml

start
:Harmonogram: Co tydzień;
:Testowe przywrócenie:\nPrzywróć losowy backup;
:Uruchom weryfikację;
:Porównaj liczniki:\n- Liczba pokemonów\n- Suma IV%\n- Unikalne UID;

if (Rozbieżności?) then (tak)
    :ALERT: Backup corrupted!;
    stop
else (nie)
endif

:Log: Test backupu - PASS;
stop

@enduml
```

> **Cel:** Upewnić się, że backupy są **grywalne** - nie tylko "istnieją".
