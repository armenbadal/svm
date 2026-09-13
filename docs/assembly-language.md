# SVM v1 assembler լեզու

Կարգավիճակ՝ **քննարկման նախագիծ 0.2**

Assembler-ը փոքր, տողային լեզու է։ Այն label-ները լուծելուց հետո կառուցում է նույն `bytecode.Program`-ը, որը կարող է անմիջապես կառուցել նաև կոմպիլյատորը։

## Source-ի ընդհանուր կանոններ

- Source-ը UTF-8 տեքստ է։
- `LF` և `CRLF` նոր տողերը համարժեք են։
- Վերջին տողը պարտադիր չէ ավարտել newline-ով։
- `;` նիշից մինչև տողի վերջը մեկնաբանություն է, եթե այն string-ի ներսում չէ։
- Mnemonic-ները, directive-ները և register անունները case-insensitive են։
- Label անունները case-sensitive են։

Identifier-ը սկսվում է Unicode տառով կամ `_`-ով և շարունակվում է տառերով, թվանշաններով կամ `_`-ով։ Օրինակ՝ `main`, `_start`, `շրջան_1`։

## Թվեր և string-եր

Աջակցվող թվային ձևերը՝

```asm
42
-17
0xff
0b1010
1_000_000
```

Hexadecimal թիվը սկսվում է `0x`-ով, binary-ն՝ `0b`-ով։ `_` բաժանիչը թույլատրվում է միայն թվանշանների միջև։ Overflow-ը assembler error է և չի կարող լուռ կտրվել։

String-ը գրվում է կրկնակի չակերտներով և UTF-8 է։ Աջակցվող escape-ներն են՝ `\\`, `\"`, `\n`, `\r`, `\t`, `\0`, `\xNN`։

```asm
"hello"
"տող\n"
```

## Grammar

```text
Program     = { Line } EOF .
Line        = [ Label ] [ Statement ] (NewLine | EOF) .
Label       = Identifier ':' .
Statement   = Instruction | Directive .
Instruction = Mnemonic [ Operand { ',' Operand } ] .
Directive   = '.' Identifier [ Operand { ',' Operand } ] .
Operand     = Number | Identifier | Address | Memory | String .
Address     = '@' Identifier [ ('+' | '-') Number ] .
Memory      = '[' Register [ ('+' | '-') Number ] ']' .
Register    = 'IP' | 'SP' | 'FP' .
```

Մեկ տողում կարող է լինել առավելագույնը մեկ label և մեկ statement։ Label-ը և instruction-ը կարող են լինել նույն տողում։

## Հրամաններ

Հիմնական շարահյուսությունը ուղիղ համապատասխանում է [`isa.md`](isa.md)-ին։

```asm
NOP
PUSH 42
PUSH -7
PUSH [FP - 12]
POP [FP + 0]
PUSH @message
DROP
DUP
SWAP
LOAD
STORE
LOADB
STOREB
ADD
SUB
MUL
DIV
MOD
NEG
AND
OR
XOR
NOT
SHL
SHR
EQ
NE
LT
LE
GT
GE
JUMP loop
JZ done
JNZ loop
CALL function
RET 2
INPUT
PRINT
HALT
```

`RET` առանց թվի համարժեք է `RET 0`-ին։ `RET n`-ում `n`-ը մաքրվող արգումենտների slot-երի քանակն է և պետք է տեղավորվի `uint16`-ում։

`PUSH @label`-ը ստեկ է դնում label-ի բացարձակ հասցեն։ Սովորական `PUSH label` ձևը չի թույլատրվում, որպեսզի հասցեն չշփոթվի label-ում պահվող արժեքի հետ։

`PUSH [register ± displacement]` և `POP [register ± displacement]` displacement-ը պետք է լինի `[-8192, 8191]` միջակայքում։ `[FP]` ձևը համարժեք է `[FP + 0]`-ին։

`CALL`, `JUMP`, `JZ` և `JNZ` ընդունում են code label։ Data label-ը որպես branch target assembler error է։

## Հատվածներ

Assembler-ն ունի միայն երկու հատված՝ code և data։

### `.code`

Ընտրում է instruction-ների հատվածը։ Սա ֆայլի սկզբնական հատվածն է։ Այստեղ թույլատրվում են instruction-ներ, label-ներ և `.align`։

### `.data`

Ընտրում է տվյալների հատվածը։ Այստեղ թույլատրվում են label-ներ և data directives։ Instruction-ը data հատվածում assembler error է։

Section switch-ը executable բայթ չի գեներացնում։ Code և data label-ները գտնվում են ընդհանուր անունների տարածքում։

## Entry point

`.entry label`-ը սահմանում է կատարման առաջին instruction-ը։ Այն կարող է գրվել մինչև label-ի սահմանումը, բայց պետք է ցույց տա code հատված։ Executable կառուցելիս ճիշտ մեկ entry point է պարտադիր։

```asm
.entry start
.code

start:
    HALT
```

## Data directives

### `.byte value, ...`

Յուրաքանչյուր արժեքից գրում է մեկ բայթ։ Թույլատրելի միջակայքը `[0, 255]` է։

```asm
.byte 0, 10, 0xff
```

### `.word value, ...`

Յուրաքանչյուր արժեքից գրում է մեկ 32-բիթանոց little-endian բառ։ Արժեքը կարող է լինել նշանով 32-բիթանոց թիվ կամ `@label` հասցե։

```asm
.word 10, -20, @message
```

### `.string "text"`

Գրում է string-ի UTF-8 բայթերը՝ առանց վերջնական զրոյի։

### `.stringz "text"`

Գրում է string-ի UTF-8 բայթերը և մեկ վերջնական `0` բայթ։

### `.zero count`

Գրում է `count` հատ զրո բայթ։ Count-ը պետք է լինի non-negative compile-time թիվ։ Առանձին BSS հատված չկա։

### `.align value`

Ընթացիկ հատվածը հավասարեցնում է տրված երկուսի աստիճանին։ Թույլատրելի արժեքներն են `1`, `2`, `4`, …, `256`։ Code հատվածի padding-ը `NOP` բայթեր են, data հատվածինը՝ զրոներ։

## Label հասցեներ

Code label-ի հասցեն իր offset-ն է code-ի սկզբից։ Data label-ի հասցեն՝

```text
align_up(code_size, 4) + data_offset
```

Forward reference-ը թույլատրվում է։ Duplicate կամ undefined label-ը assembler error է։ `@label + n` և `@label - n` ձևերը թույլատրվում են, եթե վերջնական հասցեն տեղավորվում է 16 բիթում։

## Օրինակ՝ global արժեք

```asm
.data
.align 4
counter:
    .word 0

.code
.entry start

start:
    PUSH 42
    PUSH @counter
    STORE

    PUSH @counter
    LOAD
    PRINT
    HALT
```

## Օրինակ՝ string traversal-ի հիմքը

```asm
.data
message:
    .stringz "Բարև"

.code
.entry start

start:
    PUSH @message
    LOADB
    PRINT       ; արտածում է առաջին UTF-8 բայթի թվային արժեքը
    HALT
```

Ամբողջ string-ը տպող ծրագիրը հասցեն կպահի stack/local slot-ում, `LOADB`-ով հերթով կկարդա բայթերը և կկանգնի վերջնական զրոյի վրա։ v1.0-ը հատուկ string opcode չի սահմանում։

## Diagnostics

Յուրաքանչյուր assembler error պետք է ներառի source-ի անունը, տողը և սյունը։

```text
program.asm:12:9: 'loop' պիտակը սահմանված չէ
```

Assembler-ը մերժում է՝

- անհայտ mnemonic, directive կամ register,
- operand-ների սխալ քանակ կամ տեսակ,
- թվային overflow,
- duplicate կամ undefined label,
- սխալ հատվածում instruction/directive,
- բացակայող կամ data label ցույց տվող entry point,
- data label օգտագործող branch,
- սահմաններից դուրս displacement կամ հասցե։

## Դիտավորյալ չներառված հնարավորություններ

SVM v1.0 assembler-ը չունի՝

- macros,
- `.include`,
- conditional assembly,
- `.global` և `.extern`,
- object files և linker,
- function declaration directives,
- առանձին read-only կամ BSS հատված,
- floating-point literal։
