# Glossary - Słownik Pojęć

> **Cel:** Definicje wszystkich kluczowych terminów, skrótów i wzorów używanych w systemie.

---

## A

**Account** (Konto)
: Identyfikator gracza. Plik ODS nazywany `Pokemon-GO-{account}.ods` (np. `Pokemon-GO-TEST.ods`).

**AT CAP** (Na capie)
: Status PvP Liga - pokemon na poziomie, gdzie **następny power-up przekroczy limit ligi** (1500 GL / 2500 UL). Należy **przestać levelować**.

---

## B

**Base Stats** (Statystyki bazowe)
: Wewnętrzne statystyki gatunku z Game Mastera: `atk` (atak), `def` (obrona), `hp` (HP). Podstawa do obliczania CP.

**BST** (Base Stat Total)
: Suma statystyk bazowych: `atk + def + hp`. Używane do klasyfikacji:
- `< 600` = zwykłe / common
- `600-680` = sub-legendary / pseudo-legendary
- `> 680` = box legendary / mythical

---

## C

**Cap** (Limit ligi)
: Maksymalne CP dozwolone w lidze PvP:
- **GL** (Great League) = **1500 CP**
- **UL** (Ultra League) = **2500 CP**
- **ML** (Master League) = **brak limitu**

**Candy** (Cukierki)
: Zasób do ewolucji i power-up. **Candy = rodzina, nie typ**.
- Candy Bulbasaura = tylko Bulbasaur/Ivysaur/Venusaur
- NIE = Snivy, Bellsprout, Oddish
- Każda linia ewolucji = osobna pula Candy

**Candy XL**
: Cukierki do levelowania powyżej L40. Główny bottleneck (nie Stardust). Trudniejsze do zdobycia dla legend.

**CP** (Combat Power)
: Moc bojowa pokemona. Wzór:
```
CP = floor((Atk * √Def * √HP * CPM²) / 10)
```
gdzie `Atk = base_atk + iv_atk`, itd., `CPM` = mnożnik poziomu z Game Mastera.

**CPM** (Combat Power Multiplier)
: Mnożnik poziomu z Game Mastera. 100 wartości dla poziomów 1-50 (co 0.5 lvl).
- L30 = 0.7317
- L40 = 0.8517
- L50 = 0.9717

---

## D

**Dev Score** (Score rozwoju)
: Wynik 0-100 priorytetu rozwoju. Suma 7 komponentów:
| Komponent | Max | Waga |
|-----------|-----|------|
| IV% | 15 | Jakość IV |
| Lucky | 20 | 50% Stardust discount |
| Meta Tier | 20 | S=20, A=15, B=8, C=0 |
| Moveset | 15 | Meta fast/charged |
| Mega/Shadow | 10 | Potencjał ewolucji |
| Investment | 10 | Candy XL + CP cost |
| Purpose | 10 | Rajdy/PvP=8, Trade=5, Kolekcja=3 |

---

## E

**Enrichment** (Wzbogacanie)
: Proces dodawania pól obliczonych do surowych danych z ODS: IV%, SpeciesID, Base Stats, Meta Tier, Max Safe Level, Target League, Current Level, Dev Score.

---

## F

**Force Re-released** (Wymuszone wydane)
: Gatunki oznaczone w Gamemaster jako `released: false` ale dostępne w grze (np. Ditto, Spewpa, Arrokuda, Cramorant).

---

## G

**Game Master** (Gamemaster.json)
: Źródłowy słownik danych z PvPoke: 1740 gatunków, 347 ruchów, 508 shadow-eligible, CPM table, typing, base stats, moves.

**GL** (Great League)
: Liga PvP do 1500 CP.

---

## H

**HUNDO** (100% IV)
: Pokemon z IV 100% (15/15/15). `iv_pct == 100.0`.

---

## I

**IV** (Individual Values)
: Trzy wartości indywidualne: Atak, Obrona, HP (0-15 każde). Zapis: `15/14/13` (Atk/Def/HP).
- **IVs** = trzy osobne wartości (np. `14, 13, 15`)
- **IV%** = procent sumy: `(Atk + Def + HP) / 45 * 100` (np. `93.3%`)

---

## L

**Level** (Poziom)
: Poziom pokemona (1-50, co 0.5). Każdy power-up = +0.5 lvl.
- L40 = max bez XL Candy
- L50 = max z XL Candy

**Lucky** (Lucky Pokemon)
: Pokemon uzyskany w Lucky Trade. **50% zniżka Stardust** na power-up. Wartość handlowa.

---

## M

**Max Safe Level** (Maksymalny bezpieczny poziom)
: Najwyższy poziom, na którym CP ≤ cap ligi (GL=1500, UL=2500).
Obliczany przez binary search po CPM table.

**Max CP** (Maksymalne CP)
: CP na L40/L50 przy 100% IV (15/15/15).
```
max_cp = floor((atk+15) * √(def+15) * √(hp+15) * CPM² / 10)
```

**Meta Tier** (Tier meta)
: Klasyfikacja użyteczności w meta:
- **S** = Must-have (Top 20 rankings)
- **A** = Strong (21-50)
- **B** = Niche (51-100)
- **C** = Dex filler (reszta)

---

## O

**ODS** (OpenDocument Spreadsheet)
: Format pliku `Pokemon-GO-{account}.ods`. Źródło prawdy. Edytowalne w LibreOffice.

---

## P

**PvP Liga** (PvP League)
: Liga PvP z capem CP: GL (1500), UL (2500), ML (no cap).

**PvPoke**
: Narzędzie/strona do symulacji PvP. Źródło Gamemastera i rankingsów.

**Przeznaczenie** (Purpose)
: Kolumna V w ODS: PvP, Rajdy, Gym, Trade, Transfer, Kolekcja.

**PID / UID** (Pokemon ID / Unique ID)
: Unikalny identyfikator w kolekcji: `P-XXXXX` (kolumna A w ODS). Klucz do identyfikacji konkretnego egzemplarza.

---

## S

**Score** = **Dev Score** (patrz wyżej).

**Shadow** (Shadow Pokemon)
: Pokemon z Team GO Rocket. Ma **Frustration** (zamienialny na Return podczas eventów). 20% bonus DMG, -20% DEF.

**SpeciesID** (ID gatunku)
: Unikalny identyfikator z Gamemastera (np. `mewtwo`, `charizard_mega_y`, `gyarados`).

**Stardust** (Pył gwiazd)
: Główna waluta do power-up. Lucky = 50% zniżka.

---

## T

**Target League** (Docelowa liga)
: Liga, do której celujemy z danym pokemonem. Priorytet wykrywania:
1. `goal` (kolumna Y - Cel rozwoju)
2. `purpose` (kolumna V - Przeznaczenie)
3. `tags` (tagi: Ultra League, Great League, Master League)
4. **AUTO** (wykrywanie z CP)

**Tier** = **Meta Tier** (patrz wyżej).

**Transfer** (Przeniesienie do Profesora)
: Usunięcie pokemona za 1 Candy. Niewrócone.

---

## U

**UID** = **PID** (patrz wyżej).

**UL** (Ultra League)
: Liga PvP do 2500 CP.

---

## W

**Wzór CP** (CP Formula)
```
CP = floor((Atk * √Def * √HP * CPM²) / 10)

Atk = base_atk + iv_atk
Def = base_def + iv_def
HP  = base_hp  + iv_hp
CPM = Combat Power Multiplier (z Game Mastera, zależy od poziomu)
```

---

## Z

**Zasada Złota** (Golden Rule)
: **Backup PRZED KAŻDĄ ZMIANĄ W ODS**. Ostatni backup = polisa ubezpieczeniowa.

---

## Skróty i Akronimy

| Skrót | Pełna nazwa |
|-------|-------------|
| BST | Base Stat Total |
| CP | Combat Power |
| CPM | Combat Power Multiplier |
| DPS | Damage Per Second |
| DPT | Damage Per Turn |
| EPT | Energy Per Turn |
| GL | Great League (1500 CP) |
| HUNDO | 100% IV (15/15/15) |
| IV | Individual Values |
| ML | Master League (no cap) |
| PID/UID | Pokemon ID / Unique ID |
| PvP | Player vs Player |
| TDO | Total Damage Output |
| UL | Ultra League (2500 CP) |
| XL | Candy XL (dla L40+) |

---

## Wzory Matematyczne

### CP Calculation
```
CP = floor((Atk * √Def * √HP * CPM²) / 10)

Atk = base_atk + iv_atk
Def = base_def + iv_def
HP  = base_hp  + iv_hp
```

### IV Percentage
```
iv_pct = (iv_atk + iv_def + iv_hp) / 45 * 100
```

### Max CP at Level
```
max_cp(L) = floor((base_atk+15) * √(base_def+15) * √(base_hp+15) * CPM(L)² / 10)
```

### Max Safe Level (Binary Search)
```
Find max L where CP(L) ≤ cap
CP(L) = floor((atk * √def * √hp * CPM(L)²) / 10)
```

### Dev Score
```
dev_score = 
  min(iv_pct/100 * 15, 15)                    # IV (0-15)
+ (20 if lucky else 0)                        # Lucky (0-20)
+ {"S":20, "A":15, "B":8, "C":0}[meta_tier]  # Tier (0-20)
+ moveset_score                               # Moveset (0-15)
+ mega_shadow_bonus                           # Mega/Shadow (0-10)
+ (candy_xl_score + cp_investment_score)      # Investment (0-10)
+ purpose_bonus                               # Purpose (0-10)
```

### Candy XL Feasibility
```
if BST < 600:      score = 5   # common
elif BST < 680:    score = 3   # sub-legendary
else:              score = 1   # box legendary
```

### CP Investment Score
```
cp_investment = min(5, max(0, (current_cp / max_cp_50) * 5))
```

---

## Jednostki i Precyzja

| Wartość | Zakres | Format |
|---------|--------|--------|
| IV% | 46 wartości (0/45 do 45/45) | `91.1%` |
| Level | co 0.5 lvl | `29.5` |
| CP | integer | `2496` |
| Score | 0-100, float | `86.2` |
| CPM | 6 decimal | `0.851700` |
| IVs | integer 0-15 | `14` |

---

## Mapowanie Pól ODS → Internal (Quick Reference)

| ODS Col | Letter | Internal | Type |
|---------|--------|----------|------|
| A | UID | `uid` | str |
| B | Gatunek | `species` | str |
| C | Własna nazwa | `nick` | str |
| F | CP | `cp` | int |
| G | IV Atk | `iv_a` | int |
| H | IV Def | `iv_d` | int |
| I | IV HP | `iv_h` | int |
| K | Fast Attack | `fast` | str |
| L | Charged 1 | `ch1` | str |
| M | Charged 2 | `ch2` | str |
| V | Przeznaczenie | `purpose` | str |
| Y | Cel rozwoju | `goal` | str |

---

## Flagi Pokemonów (Status Flags)

| Flaga | Źródło | Znaczenie |
|-------|--------|-----------|
| `is_shiny` | tags | Shiny variant |
| `is_lucky` | tags | Lucky trade (50% dust) |
| `is_shadow` | tags | Shadow (Frustration) |
| `is_purified` | tags | Purified (Return) |
| `is_dynamax` | tags | Dynamax capable |
| `is_traded` | tags | Traded (XL candy) |
| `is_hundo` | computed | IV% == 100% |
| `is_mega` | species_id | Mega evolution |