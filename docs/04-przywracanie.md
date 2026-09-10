# Przywracanie ODS z Backupu

> **Cel:** Procedury odzyskiwania danych z backupów - od prostego cofnięcia po disaster recovery.

---

## Scenariusze Przywracania

```plantuml
@startuml

start
:Potrzeba przywrócenia;
note right
  Scenariusze:
  - ODS uszkodzony/nieotwieralny
  - Zła edycja w LibreOffice
  - Regeneracja z błędnymi danymi
  - Przypadkowe usunięcie ODS
  - Chcesz wersję z X dni temu
  - Totalna awaria dysku/systemu
end note

switch (Rodzaj przywracania)
case (Prosty przypadek)
    :Przywróć najnowszy backup;
case (Wybrana wersja)
    :Wybierz konkretny backup;
case (Awaria)
    :Odtwórz z zewnętrznego nośnika;
endswitch

:Weryfikacja: sprawdz + LO;
:Wznowienie pracy;
stop

@enduml
```

---

## Komenda Przywracania

```plantuml
@startuml

start
:Przywróć ODS;
:Argumenty: ścieżka backupu [docelowa ścieżka ODS];

:Rozwiąż backup:\n- Pełna ścieżka LUB\n- Nazwa pliku [szukaj w backups/];

if (Znaleźiono plik?) then (tak)
else (nie)
    :FAIL: Backup nie znaleziony;
    stop
endif

:Weryfikuj backup:\n- ZIP valid\n- content.xml exists\n- size > 100B;

if (Weryfikacja OK?) then (tak)
else (nie)
    :FAIL: Backup uszkodzony;
    stop
endif

:Rozwiąż target ODS:\n- Podana ścieżka LUB\n- Domyślny: Pokemon-GO-{X}.ods\n  (wyciągnięty z nazwy backupu);

:Kopiuj backup → target\nNadpisuje istniejący!;

:Weryfikuj przywrócony;\n[logika weryfikacji];

if (Weryfikacja OK?) then (tak)
else (nie)
    :FAIL: Przywrócony uszkodzony;
    stop
endif

:SUCCESS/FAIL + ścieżka;
stop

@enduml
```

---

## Interaktywne Wybieranie Backupu

```plantuml
@startuml

start
:Uruchomienie bez argumentów;

:Wylistuj backups/\nDla domyślnego konta LUB podanego;

:Wyświetl menu:\n1. Pokemon-GO-{konto}.ods.bak.20260906_080000\n2. Pokemon-GO-{konto}.ods.bak.20260905_112000\n...\n0. Anuluj;

while (User wybiera numer) is (podany)
    if (Poprawny numer?) then (tak)
        :Wybrany backup;
    else (nie)
        :Pokaż błąd;
    endif
endwhile (wybrano)

if (Anuluj?) then (tak)
    stop
endif

:Potwierdź: Nadpisać ODS?\n[t/N];

if (Potwierdzono?) then (tak)
    :Wykonaj przywracanie;
else (nie)
    stop
endif

stop

@enduml
```

---

## Procedury Specyficzne

### 1. Cofnięcie złej edycji w LibreOffice

```plantuml
@startuml

start
:Zła edycja w LO:\n- Usunięto pokemony\n- Zmieniono IV/CP\n- Zepsuto formuły;

:Zamknij LibreOffice\n[NIE ZAPISUJ!];
:Backup obecnego stanu [opcjonalnie];
:Przywróć najnowszy backup;
:Weryfikacja;
:Otwórz w LO\nSprawdź dane;
:EDYCJA COFNIĘTA;
stop

@enduml
```

> **Wskazówka:** Jeśli nie zapisałeś w LO - plik ODS na dysku jest nienaruszony. Backup potrzebny tylko jeśli już zapisałeś.

### 2. Regeneracja ODS z błędnymi danymi

```plantuml
@startuml

start
:Regeneracja wygenerowała\nzłe dane:\n- Złe VLOOKUP\n- Brak arkuszy\n- Złe walidacje;

:Przywróć backup sprzed regeneracji;
:Weryfikacja;
:Otwórz w LO\nSprawdź dane;
:Popraw dane w ODS\nLUB napraw logikę regeneracji;
:Ponowna regeneracja;
:ODS NAPRAWIONY;
stop

@enduml
```

### 3. Przywrócenie do wersji z X dni temu

```plantuml
@startuml

start
:Chcesz wersję z dnia X;
:Wylistuj backupy [bez argumentów];
:Zidentyfikuj backup:\n20260903 = 3 września;

if (Potwierdź datę?) then (tak)
else (nie)
    :Powrót do listy;
    stop
endif

:Przywróć wybrany backup;
:Weryfikacja + LO check;
:Wersja z X przywrócona;
stop

@enduml
```

### 4. Disaster Recovery (Totalna awaria)

```plantuml
@startuml

start
:Totalna awaria:\n- Uszkodzony dysk\n- Utracone ODS + backups lokalne;

if (Masz backupy\nzewnętrzne?\n[USB, chmura, NAS]) then (tak)
else (nie)
    :STRATA DANYCH - BRAK BACKUPU ZEWNĘTRZNEGO;
    stop
endif

:Skopiuj backups/\nna nowy system/dysk;
:Przywróć z backupu;
:Weryfikacja;
:Przywróć skrypty\n[git clone / kopia];
:Przywróć data/gamemaster.json\n[git / pobierz];
:SYSTEM ODZYSKANY;
stop

@enduml
```

> **KRITYCZNE:** Backupy muszą być **zewnętrzne** (USB, chmura, NAS). Lokalne backups/ na tym samym dysku nie chronią przed awarią dysku.

---

## Weryfikacja Po Przywróceniu (Checklista)

```plantuml
@startuml

start
:ODS przywrócony;
:Uruchom weryfikację => SUCCESS?;

if (Weryfikacja OK?) then (tak)
else (nie)
    :FAIL;
    stop
endif

:Otwórz w LibreOffice => Otwiera się?;

:Sprawdź arkusz Kolekcja:\n- Liczba wierszy = oczekiwana?\n- UID kluczowe obecne?\n- Walidacje działają [dropdowny]?;

:Sprawdź arkusz Pokedex:\n- AutoSync checkboxy OK?\n- 1598 wierszy?;

:Porównaj z JSON snapshot:\n- Liczba pokemonów\n- Suma IV%\n- Unikalne UID\n- Suma CP;

if (Wszystko OK?) then (tak)
    :CONTINUITY CONFIRMED;
else (nie)
    :Rozwiń różnice\nprzed dalszą pracą;
endif

stop

@enduml
```

> **Zasada:** Jeśli jakakolwiek metryka się nie zgadza → **STOP**, nie pracuj dalej, rozwiąż różnicę.

---

## Automatyzacja Weryfikacji (Post-Restore Hook)

```plantuml
@startuml

start
:Po przywracaniu;
:Uruchom automatyczną weryfikację;

if (Wynik?) then (FAIL)
    :ALERT: Backup corrupted!;
    stop
else (PASS)
endif

:Wygeneruj mini-raport:\n- Timestamp przywracania\n- Źródło backupu\n- Wynik weryfikacji\n- Metryki porównawcze;
:Zapisz do logs/restore.log;
:Opcjonalnie: powiadomienie\n[terminal / plik];
:Restore + Verify Complete;
stop

@enduml
```

---

## Logowanie Przywracania

```plantuml
@startuml

left to right direction

rectangle "restore.log" {
    [2026-09-06 10:15\nRESTORE\nkonto.bak.20260905 => konto\nSUCCESS\nverify=PASS\npokemons=670] as L1
    [2026-09-06 10:16\nRESTORE\nkonto2.bak.20260903 => konto2\nSUCCESS\nverify=PASS\npokemons=708] as L2
    [2026-09-06 10:18\nRESTORE\nkonto.bak.20260901 => konto\nFAIL\nverify=FAIL\nreason=corrupt_zip] as L3
}

note bottom
  Format: Timestamp | RESTORE | source => target | STATUS | verify=PASS/FAIL | metrics
end note

@enduml
```

---

## Integracja z Backup Policy

```plantuml
@startuml

|Backup Policy|
:Max 5 kopii, weryfikowane;

|Backup|
:Tworzy weryfikowany plik\nw backups/ z timestampem;

|Prune|
:Zachowuje 5 najnowszych\n= 5 punktów przywracania;

|Restore|
:Wymaga istniejącego backupu\nZweryfikowanego backupu;

|Cycle|
:Backup => [czas] => Awaria =>\nRestore => Verify => Kontynuacja;

@enduml
```

> **Zasada:** System backup-restore jest **domknięty** - backup tworzy plik, który restore potrafi odtworzyć, a verify potwierdza.
