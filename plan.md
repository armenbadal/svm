# SVM v1-ի զարգացման պլան

Նպատակն է կառուցել հստակ սահմանված **SVM v1**, որը հնարավոր կլինի օգտագործել և՛ որպես ինքնուրույն վիրտուալ մեքենա, և՛ որպես կոմպիլյատորի նպատակային մեքենա։ Հիմքը պետք է լինի library-first ճարտարապետություն, որի վրա նույն կերպ կաշխատեն ասեմբլերը, արտաքին կոմպիլյատորը և CLI-ն։

## Թիրախային ճարտարապետությունը

```text
Assembler source ──> assembler ──┐
                                 ├──> bytecode.Builder ──> Program ──> Encoder
Compiler AST/IR ─────────────────┘                              │
                                                               ▼
.svm executable ──> Decoder ──> Verifier ──> VM ──> Host I/O / Result
```

Առաջարկվող փաթեթները՝

- `isa` — հրամանների հավաքածուի միակ հեղինակավոր նկարագրությունը։
- `bytecode` — հրամանների կոդավորում, ապակոդավորում և ծրագրի ձևաչափ։
- `assembler` — տեքստից `bytecode.Program`։
- `linker` — նշանների և relocation-ների լուծում։
- `vm` — մեկուսացված կատարման շարժիչ։
- `cmd/svm` — բարակ հրամանային ինտերֆեյս։
- `debug` — disassembler, trace և breakpoint-ներ։
- `docs` — ISA, ABI, assembler և binary format specification։

## Փուլ 0․ ներկա վիճակի կայունացում

Նախ անհրաժեշտ է ավարտել կամ համակարգել արդեն առկա չպահպանված փոփոխությունները։

- Վերականգնել բոլոր թեստերի անցումը։
- Ուղղել assembler-ի ներկայիս panic-ը `assembler/parser_test.go` ֆայլում։
- Սահմանել ներկա բայթ-կոդի golden fixture-ներ, որպեսզի հետագա փոփոխությունները տեսանելի լինեն։
- Առանձնացնել պատահական սխալների ուղղումները ճարտարապետական վերափոխումներից։
- Ներկա v0 ձևաչափը փաստագրել այնքանով, որքանով պետք է migration-ը հասկանալու համար։

Ավարտի չափանիշ՝ `go test ./...` ամբողջությամբ անցնում է։

## Փուլ 1․ SVM v1 specification

Նախ կոդից անկախ պետք է սահմանել մեքենայի պայմանագիրը։

Առաջարկվող հիմնական որոշումները՝

- արժեք՝ նշանով 32-բիթանոց ամբողջ թիվ,
- հասցե՝ 32-բիթանոց աննշան ամբողջ թիվ,
- հիշողություն՝ byte-addressed,
- byte order՝ little-endian,
- բառի չափ՝ 4 բայթ,
- ստեկն աճում է դեպի մեծ հասցեներ,
- 32-բիթանոց բեռնումներն ու գրառումները պահանջում են 4-բայթանոց հավասարեցում,
- թվաբանական overflow-ը սահմանվում է որպես 32-բիթանոց wraparound,
- սխալ կատարումը հանգեցնում է typed trap-ի, ոչ թե Go `panic`-ի։

Պետք է ստեղծվեն առնվազն՝

- `docs/isa.md`,
- `docs/memory.md`,
- `docs/abi.md`,
- `docs/bytecode-format.md`,
- `docs/assembly-language.md`։

Քանի որ ներկա ձևաչափը չվերսիոնավորված և դեռ անկայուն է, SVM v1-ում թույլատրվում է վերահսկվող անհամատեղելի փոփոխություն։

## Փուլ 2․ ISA-ի միասնական մոդել

Այս պահին opcode-ների մասին գիտելիքը բաշխված է տարբեր փաթեթներում։ Պետք է ունենալ մեկ աղյուսակ, որը յուրաքանչյուր հրամանի համար սահմանում է՝

- opcode,
- mnemonic,
- operand-ների տեսակները,
- instruction size,
- ստեկի ազդեցությունը,
- թույլատրելի addressing mode-երը,
- հնարավոր trap-երը։

Նվազագույն ամբողջական ISA-ն պետք է ներառի՝

- ստեկ՝ `PUSH`, `DROP`, `DUP`, `SWAP`, `SLIDE`,
- հիշողություն՝ `LOAD8`, `LOAD32`, `STORE8`, `STORE32`, `ADDR`,
- թվաբանություն՝ `ADD`, `SUB`, `MUL`, `DIV`, `MOD`, `NEG`,
- բիթային գործողություններ՝ `AND`, `OR`, `XOR`, `NOT`, `SHL`, `SHR`,
- համեմատումներ՝ `EQ`, `NE`, `LT`, `LE`, `GT`, `GE`,
- կառավարում՝ `JUMP`, `JZ`, `JNZ`, `CALL`, `RET`, `HALT`,
- host ծառայություն՝ `SYSCALL` կամ համարժեք trap interface։

`INPUT` և `PRINT` հրամանները կարելի է պահպանել որպես assembler shorthand, բայց VM-ի հիմքում I/O-ն պետք է լինի host abstraction, որպեսզի ներկառուցվող VM-ը կախված չլինի `os.Stdin`-ից և `fmt.Println`-ից։

## Փուլ 3․ ծրագրի և binary format-ի ստեղծում

Հում `[]byte`-ի փոխարեն անհրաժեշտ է կառուցվածքային `Program`։

Այն պետք է ունենա՝

- magic number,
- format version,
- entry point,
- code section,
- read-only data section,
- mutable data section,
- BSS չափ,
- պահանջվող հիշողության չափ,
- optional symbol table,
- optional source/debug metadata։

Հիմնական ֆայլային ձևաչափերը՝

- `.svm` — կատարվող ծրագիր,
- `.svo` — relocatable object file, երբ ավելացվի բազմամոդուլ կոմպիլյացիան։

Decoder-ը պետք է մերժի սխալ version-ը, կտրված հրահանգները, անվավեր opcode-ները և section-ների սահմաններից դուրս հասցեները։

## Փուլ 4․ կոմպիլյատորի համար հանրային API

Կոմպիլյատորը չպետք է ստիպված լինի տեքստային assembler գեներացնել։ Այն պետք է կարողանա անմիջապես կառուցել ծրագիր։

Նախատեսվող API-ի գաղափարը՝

```go
builder := bytecode.NewBuilder()
main := builder.NewLabel("main")

builder.Mark(main)
builder.PushInt(42)
builder.Halt()

program, err := builder.Build(bytecode.BuildOptions{
    Entry: main,
})
```

API-ն պետք է ապահովի՝

- typed instruction emission,
- labels և relocations,
- code/data sections,
- constants և string literals,
- globals,
- source locations,
- symbol/debug metadata,
- կառուցման ժամանակ operand range validation,
- undefined և duplicate symbol diagnostics։

Ներկայիս `Validate() bool` մոտեցումը պետք է փոխարինվի իմաստալից `error` կամ diagnostics ցուցակով։

## Փուլ 5․ assembler-ի ամբողջականացում

Assembler-ը պետք է դառնա նույն builder-ի frontend-ը։

Անհրաժեշտ հնարավորությունները՝

- `io.Reader`-ից և string-ից assemble անող API,
- ֆայլային wrapper միայն CLI մակարդակում,
- հստակ source position՝ ֆայլ, տող, սյուն,
- թվեր՝ տասնորդական, hexadecimal և binary,
- string և character literals,
- label expressions՝ `label + 4`,
- մեկնաբանություններ,
- case-sensitivity-ի հստակ կանոն,
- sections և directives։

Նախնական directives՝

```asm
.entry main
.code
.data
.word 42
.byte 0xff
.string "hello"
.zero 64
.align 4
.global main
.extern print_i32
```

Assembler-ը պարտադիր պետք է հայտնաբերի՝

- անհայտ mnemonic,
- պակասող կամ ավելորդ operand,
- սխալ addressing mode,
- թվային overflow,
- կրկնված label,
- չսահմանված symbol,
- սխալ section օգտագործում։

Ցանկալի է իրականացնել lexer → parser/AST → semantic validation → emission շղթայով, ոչ թե անմիջապես token-ներից bytecode գրելով։

## Փուլ 6․ ABI և կանչերի պայմանագիր

Կոմպիլյատորի համար սա ամենակարևոր պայմանագիրն է։

Պետք է վերջնականապես սահմանել՝

- արգումենտների տեղադրման հերթականությունը,
- վերադարձվող արժեքների քանակը,
- caller-saved/callee-saved վիճակը,
- `FP`-ի նշանակությունը,
- return address-ի և նախորդ `FP`-ի դիրքերը,
- լոկալների դասավորությունը,
- ստեկի մաքրող կողմը,
- ծրագրի entry point-ի պայմանագիրը։

Առաջարկվող տարբերակ՝

- բոլոր արժեքները մեկ 32-բիթանոց slot են,
- caller-ը հերթով push է անում արգումենտները,
- `CALL`-ը պահպանում է return address-ը և նախորդ `FP`-ը,
- ֆունկցիան միշտ վերադարձնում է մեկ slot,
- `RET`-ից հետո caller-ը `SLIDE n`-ով հեռացնում է արգումենտները՝ պահպանելով արդյունքը,
- void արդյունքը ներկայացվում է `0`-ով։

Պետք է ունենալ առանձին օրինակներ recursion-ի, nested call-ի և մի քանի արգումենտով ֆունկցիայի համար։

## Փուլ 7․ VM runtime-ի վերակառուցում

Ներկա runtime-ն անմիջապես աշխատում է ներքին հիշողության, `fmt`-ի և panic-ների հետ։ Նոր runtime-ը պետք է ունենա այսպիսի օգտագործման ձև՝

```go
machine, err := vm.New(vm.Config{
    MemorySize: 1 << 20,
    Host:       host,
    MaxSteps:   1_000_000,
})

result, err := machine.Run(ctx, program)
```

Պահանջվող հնարավորությունները՝

- `Load`, `Reset`, `Run`, `Step`,
- configurable memory,
- custom host I/O,
- execution step limit,
- `context.Context`-ով ընդհատում,
- typed traps,
- machine state-ի read-only snapshot,
- deterministic execution,
- stack underflow/overflow պաշտպանություն,
- instruction fetch bounds,
- alignment checks,
- code/data սահմանների ստուգում,
- division և modulo by zero traps,
- invalid opcode/mode traps։

Հանրային API-ի սովորական սխալներից ոչ մեկը չպետք է ավարտվի Go panic-ով։

## Փուլ 8․ verifier և debugger գործիքներ

Verifier-ը մինչև կատարումը պետք է ստուգի՝

- instruction boundaries,
- jump/call target-ներ,
- operand encoding,
- entry point,
- section limits,
- ակնհայտ stack-effect անհամապատասխանություններ։

Debugger-ի առաջին տարբերակը՝

- instruction trace,
- register dump,
- stack dump,
- breakpoints,
- step/continue,
- disassembler,
- symbol անունների ցուցադրում։

Սա միաժամանակ շատ օգտակար կլինի կոմպիլյատորից գեներացված սխալ կոդը հետազոտելու համար։

## Փուլ 9․ CLI

Մեկ `svm` executable՝ ենթահրամաններով։

```sh
svm asm program.asm -o program.svm
svm run program.svm
svm exec program.asm
svm check program.svm
svm disasm program.svm
svm debug program.svm
svm version
```

Կարևոր դրոշներ՝

- `--memory`,
- `--max-steps`,
- `--trace`,
- `--entry`,
- `--strip`,
- `--debug-info`։

CLI-ն պետք է ունենա կանխատեսելի exit code-եր և stdout/stderr-ի հստակ տարանջատում։

## Փուլ 10․ թեստավորում և որակի շեմ

Թեստային շերտերը՝

- յուրաքանչյուր ISA հրամանի unit test,
- encoder/decoder round-trip,
- assembler golden tests,
- malformed input tests,
- VM trap tests,
- ABI contract tests,
- assembler-vs-builder bytecode equivalence,
- end-to-end ծրագրեր,
- fuzz tests՝ lexer, parser, decoder և verifier,
- deterministic output tests։

Պարտադիր end-to-end օրինակներ՝

- թվաբանական հաշվարկ,
- պայման և ցիկլ,
- ֆունկցիայի կանչ,
- recursion,
- globals,
- array կամ string memory access,
- host I/O,
- կոմպիլյատորի API-ով կառուցված նույն ծրագրի կատարում։

## Փուլ 11․ փաստաթղթեր և թողարկում

- README-ն դարձնել արագ մեկնարկի ուղեցույց։
- Ավելացնել assembler reference։
- Ավելացնել compiler backend integration guide։
- Նկարագրել ABI-ն և executable format-ը։
- Սահմանել semantic versioning։
- Եթե փաթեթը հրապարակվելու է, `module svm`-ը փոխարինել կայուն repository module path-ով։
- Ավելացնել CI՝ formatting, tests, race detector և cross-platform build ստուգումներով։

## Առաջին թողարկման սահմանները

SVM v1-ի համար նպատակահարմար է չներառել՝

- JIT,
- garbage collector,
- threads,
- floating-point,
- dynamic linking,
- ուղիղ filesystem կամ network հասանելիություն։

Դրանք կարելի է ավելացնել հետագայում՝ առանց հիմնական ISA/ABI-ն կոտրելու։ Առաջին լիարժեք տարբերակի շեմը պետք է լինի անվտանգ, deterministic և debug-friendly 32-բիթանոց VM, որը կարող է կատարել իրական կոմպիլյատորի գեներացրած ֆունկցիաներ, globals, ցիկլեր, պայմաններ և պարզ տվյալային կառուցվածքներ։

## Իրականացման հերթականությունը

```text
baseline
  → specification
  → ISA և binary format
  → builder API
  → assembler
  → ABI և VM runtime
  → CLI և debugger
  → fuzzing, փաստաթղթեր և թողարկում
```

Յուրաքանչյուր փուլ պետք է ավարտվի անցնող թեստերով և առանձին ստուգելի արդյունքով՝ առանց հաջորդ փուլի անավարտ կոդից կախված լինելու։
