# Akcje Użytkownika - Context Diagram

> **Cel:** Przedstawić wszystkie akcje jakie może wykonać Aktor (Operator) na systemie, bez wiązania do konkretnych plików/skryptów.
> System jest agnostyczny względem implementacji (Python, Go, Rust, itp.).

---

## Context Overview

```plantuml
@startuml

package "External Systems" {
    [PvPoke GitHub\nMeta Rankings] as PvPoke
    [Game Master\nCPM, Base Stats, Moves] as GameMaster
}

package "System Boundary" {
    actor "Operator / Użytkownik" as Actor
    
    package "Commands / Actions" {
        [Generuj ODS\nDla konta X] as GenerateODS
        [Aktualizuj ODS\nNowe pokemony] as UpdateODS
        [Backup ODS\nZ weryfikacją] as BackupODS
        [Weryfikuj ODS\nStruktura + dane] as VerifyODS
        [Przywróć ODS\nZ backupu] as RestoreODS
        [Analizuj Konto\nPełny raport] as Analyze
        [Aktualizuj Meta\nRankingi PvPoke] as UpdateMeta
    }
    
    package "Data Stores" {
        [Collection ODS\nPokemon-GO-{account}.ods] as CollectionODS
        [Gamemaster JSON\nBase stats, moves, CPM] as GamemasterJSON
        [Meta Rankings JSON\nPvPoke tiers S/A/B/C] as MetaRankingsJSON
        [Backups\nTimestamped copies] as BackupsDir
        [Analizy\nanalizy/Analiza-{date}-{model}-{account}.md] as ReportsDir
        [Collection JSON\nanalizy/collection-{account}-{date}.json] as CollectionJSON
    }
}

Actor --> GenerateODS : wywołuje
Actor --> UpdateODS : wywołuje
Actor --> BackupODS : wywołuje
Actor --> VerifyODS : wywołuje
Actor --> RestoreODS : wywołuje
Actor --> Analyze : wywołuje
Actor --> UpdateMeta : wywołuje

GenerateODS --> CollectionODS : reads + writes
GenerateODS --> GamemasterJSON : reads

UpdateODS --> CollectionODS : reads + writes

BackupODS --> CollectionODS : reads
BackupODS --> BackupsDir : writes

VerifyODS --> CollectionODS : reads
VerifyODS --> ReportsDir : outputs

RestoreODS --> BackupsDir : reads
RestoreODS --> CollectionODS : writes

Analyze --> CollectionODS : reads
Analyze --> GamemasterJSON : reads
Analyze --> MetaRankingsJSON : reads
Analyze --> ReportsDir : writes
Analyze --> CollectionJSON : writes

UpdateMeta --> PvPoke : fetches
UpdateMeta --> MetaRankingsJSON : writes

GamemasterJSON ..> GameMaster : sourced from
MetaRankingsJSON ..> PvPoke : sourced from

@enduml
```

---

## Akcje - Szczegółowe Widoki

### 1. Generuj ODS (Create New Collection)

```plantuml
@startuml

|Actor|
start
:Generuj ODS dla konta X;

|System|
if (Collection ODS\nistnieje?) then (tak)
else (nie)
endif
:Load Gamemaster JSON\n[base stats, moves, CPM];
:Stwórz szablon ODS\nArkusze: Kolekcja, Pokedex,\nPokemon, Ruchy, Listy, itp.;
:Wypełnij Pokedex\n1598 gatunków z Gamemaster\n+ AutoSync z kolekcji;
:Zapisz Pokemon-GO-{X}.ods;

|Actor|
:Gotowy ODS;
stop

@enduml
```

### 2. Aktualizuj ODS (Add New Pokemon to Existing)

```plantuml
@startuml

|Actor|
start
:Aktualizuj ODS;

|System|
:Load Collection ODS\n[Arkusz: Kolekcja];
:Load Gamemaster JSON;

|Actor|
:Użytkownik edytuje w LibreOffice\nDodaje nowe wiersze w Kolekcja;
:Zapisuje ODS;

|System|
if (Czy regenerować\narkusze pochodne?) then (tak)
    :Regeneruj Pokedex, Pokemon,\nRuchy, Listy na bazie Kolekcja;
    :Zapisz zaktualizowane ODS;
else (nie)
endif

|Actor|
:Zaktualizowane ODS;
stop

@enduml
```

### 3. Backup ODS (z weryfikacją)

```plantuml
@startuml

|System|
start
:Backup ODS;

:Odczytaj Collection ODS;
:Wygeneruj timestamp\nYYYYMMDD_HHMMSS;
:Skopiuj do backups/\nPokemon-GO-{X}.ods.bak.{timestamp};
:Zweryfikuj backup:\n- ZIP valid\n- content.xml exists\n- size > 100B;

if (Zweryfikuj backup?) then (FAIL)
    :Log: FAIL + cleanup;
    stop
else (OK)
endif

if (Usuń stare\n[max 5 kopii]?) then (tak)
    :Usuń najstarsze\npowyżej limitu 5;
else (nie)
endif

:Log: SUCCESS + ścieżka backupu;

|Actor|
:Backup gotowy;
stop

@enduml
```

### 4. Weryfikuj ODS

```plantuml
@startuml

|System|
start
:Weryfikuj ODS;

:Otwórz jako ZIP;

if (content.xml\nistnieje?) then (tak)
else (nie)
    :FAIL: brak content.xml;
    stop
endif

:Parsuj content.xml;
:Policz arkusze;
:Policz wiersze [min 100];
:Policz komórki [min 1000];
:Sprawdź nazwy arkuszy;
:Log: SUCCESS/FAIL\n+ statystyki;

|Actor|
:Raport weryfikacji;
stop

@enduml
```

### 5. Przywróć ODS z Backupu

```plantuml
@startuml

|Actor|
start
:Przywróć ODS;
:Wybierz backup\n[nazwa pliku lub pełna ścieżka];

|System|
if (Znaleźiono\nw backups?) then (tak)
else (nie)
    :FAIL: nie znaleziono;
    stop
endif

:Zweryfikuj backup\n[ZIP + content.xml];

if (Zweryfikuj backup?) then (FAIL)
    :FAIL: backup uszkodzony;
    stop
else (OK)
endif

:Skopiuj jako\nPokemon-GO-{X}.ods;
:Zweryfikuj przywrócony plik;
:Log: SUCCESS/FAIL;

|Actor|
:ODS przywrócony;
stop

@enduml
```

### 6. Analizuj Konto (Pełny Raport)

```plantuml
@startuml

|System|
start
:Analizuj Konto X;

:Load Collection ODS\n[Arkusz: Kolekcja];
:Load Gamemaster JSON;
:Load Meta Rankings JSON\n[PvPoke tiers];

:Enrichment:\n- Oblicz IV%\n- Wykryj Shiny/Lucky/Shadow\n- Mapuj na SpeciesID\n- Pobierz Base Stats, Types, Moves\n- Oblicz Max Safe Level per Cap\n- Wykryj Target League [goal/purpose];

:Generuj Sekcje 1-6:\n1. Podsumowanie\n2. Rajdy Tier List\n3. PvP Teamy\n4. Transfer\n5. Lucky Trade\n6. Szczegółowa Analiza;

:Pipeline Sekcja 7:\nFiltry => Scoring => League Cap => Sort => Output;

:Renderuj Markdown Report;
:Zapisz analizy/Analiza-{date}-opencode-{X}.md;
:Zapisz analizy/collection-{X}-{date}.json;

|Actor|
:Raport gotowy;
stop

@enduml
```

### 7. Aktualizuj Meta Rankings (PvPoke)

```plantuml
@startuml

|System|
start
:Aktualizuj Meta;

if (Cache < 30 dni?) then (tak - aktualny)
    :Meta zaktualizowane;
    stop
else (nie / --force)
endif

:Pobierz z GitHub PvPoke:\n- Great League Overall/Leads/Switches/Closers\n- Ultra League Overall/...\n- Master League Overall/...\n- Raid Attackers;

if (Pobierz z GitHub?) then (FAIL)
    :Użyj hardcoded fallback;
else (OK)
    :Parsuj JSON rankings\n[speciesId + rating];
endif

:Przypisz Tiery:\nTop 20 = S, 21-50 = A, 51-100 = B, reszta = C;
:Dopełnij hardcoded\nfallback tiers S/A/B;
:Zapisz data/meta_rankings.json\n+ timestamp;

|Actor|
:Meta zaktualizowane;
stop

@enduml
```

---

## Podsumowanie Encji i Relacji

| Encja | Źródło | Opis |
|-------|--------|------|
| **Collection** | ODS (Arkusz: Kolekcja) | Wiersze = pokemony użytkownika (UID, IV, CP, moves, tags, purpose, goal) |
| **Gamemaster** | JSON (PvPoke) | Base stats, moves, CPM table, shadow/mega eligible, typing |
| **MetaRankings** | JSON (cached PvPoke) | Tiery S/A/B/C per speciesId |
| **Backup** | Timestamped ODS copy | Max 5 kopii, weryfikowane |
| **Report** | Markdown | 7-sekcja analiza + ranking rozwoju |
| **CollectionJSON** | JSON snapshot | Debug / dalsze przetwarzanie |
