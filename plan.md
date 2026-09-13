# SVM v1-ի զարգացման պլան

Նպատակն է կառուցել փոքր և ամբողջական ուսումնական ստեկային վիրտուալ մեքենա, որը հնարավոր կլինի օգտագործել և՛ հրամանային տողից, և՛ որպես ուսումնական կոմպիլյատորի նպատակային մեքենա։

Մեքենան պետք է այնքան պարզ լինի, որ դրա հիշողությունը, instruction encoding-ը, call frame-ը և fetch–decode–execute ցիկլը հնարավոր լինի ամբողջությամբ բացատրել առանց production runtime-ի լրացուցիչ շերտերի։

## Թիրախային կառուցվածքը

```text
Assembler source ──> assembler ──┐
                                 ├──> bytecode.Builder ──> Program ──> .svm
Compiler AST/IR ─────────────────┘                              │
                                                               ▼
                                                     Machine.Load/Run
```

Հիմնական փաթեթները՝

- `bytecode` — opcode-ներ, instruction encoding, `Program` և `Builder`,
- `assembler` — տեքստային ծրագրից `Program`,
- `machine` — հիշողություն, երեք ռեգիստր և կատարման ցիկլ,
- `cmd/svm` — assemble/run հրամանային գործիք։

Առանձին linker, object format, syscall layer կամ function metadata առաջին տարբերակում չի նախատեսվում։

## Փուլ 0․ ներկա վիճակի կայունացում

Կարգավիճակ՝ **ավարտված**։

- Թեստային հավաքածուն կայունացված է։
- Parser-ը մերժում է սխալ operand-ները, overflow-ը և կրկնված label-ները։
- Builder-ը հայտնաբերում է չսահմանված label-ները։
- Հարաբերական բացասական հասցեավորումն ուղղված է։
- VM-ի հիշողության, ստեկի և call frame-ի սահմանները ստուգվում են։
- Ավելացված են golden, malformed-input, end-to-end և recursion թեստեր։
- v0 format-ը նկարագրված է [`docs/v0-bytecode.md`](docs/v0-bytecode.md)-ում։

## Փուլ 1․ SVM v1 specification

Կարգավիճակ՝ **պատրաստ է վերանայման**։

Քննարկման նախագիծը սկսվում է [`docs/specification.md`](docs/specification.md) ինդեքսից։

Հիմնական առաջարկները՝

- արժեք՝ նշանով 32-բիթանոց ամբողջ թիվ,
- static հասցե՝ 16 բիթ,
- հիշողություն՝ առավելագույնը 64 ԿԲ, default՝ 16 ԿԲ,
- մեկ byte-addressed little-endian հիշողություն,
- `IP`, `SP`, `FP` երեք ռեգիստր,
- code → data → stack պարզ դասավորություն,
- ներկայիս `00/01/10` instruction encoding-ի պահպանում,
- լոկալների հատկացում սովորական `PUSH 0`-ներով,
- callee cleanup՝ `RET n`,
- `INPUT`/`PRINT` իրական opcode-ներ՝ ներարկվող I/O հոսքերով,
- 12-բայթանոց versioned executable header,
- պարզ `RuntimeError`՝ առանց մեծ trap hierarchy-ի։

Specification-ի հաստատումից առաջ implementation-ը չի սկսվում։

## Փուլ 2․ v1 instruction encoding և Program

- Սահմանել opcode/mode-ի մեկ հեղինակավոր աղյուսակ։
- Պահպանել v0-ի 0–24 opcode արժեքները։
- Ավելացնել `DROP`, `DUP`, `SWAP`, `XOR`, `SHL`, `SHR`, `JNZ`, `LOAD`, `STORE`, `LOADB`, `STOREB`։
- `RET`-ի Short ձևում կոդավորել մաքրվող արգումենտների քանակը։
- Իրականացնել անվտանգ instruction encoder/decoder։
- Ավելացնել կառուցվածքային `Program`՝ `Entry`, `Code`, `Data` դաշտերով։
- Decoder-ում ստուգել mode-ը, operand-ի ամբողջականությունը և instruction boundary-ները։

Ավարտի չափանիշ՝ բոլոր instruction-ների encode/decode golden և round-trip թեստերը կանաչ են։

## Փուլ 3․ compiler-facing Builder

Կոմպիլյատորը չպետք է assembler text գեներացնելու կարիք ունենա։

Առաջարկվող օգտագործումը՝

```go
builder := bytecode.NewBuilder()
main := builder.Label("main")

builder.Code()
builder.Mark(main)
builder.PushInt(42)
builder.Print()
builder.Halt()

program, err := builder.Build(bytecode.BuildOptions{
    Entry: main,
})
```

Builder-ը պետք է ապահովի՝

- typed instruction emission,
- code/data հատվածների ընտրություն,
- code և data label-ներ,
- forward reference-ներ,
- թվեր, բայթեր և string տվյալներ,
- entry point,
- operand range validation,
- duplicate և undefined label diagnostics,
- deterministic `Program`։

Ավարտի չափանիշ՝ Builder-ով կառուցված factorial/global/array ծրագրերը աշխատում են առանց assembler-ի։

## Փուլ 4․ assembler-ի ամբողջականացում

Assembler-ը պետք է դառնա Builder-ի փոքր տեքստային frontend-ը։

Ավելացնել՝

- `io.Reader` և string source API,
- source file/line/column diagnostics,
- decimal, hexadecimal և binary թվեր,
- `@label` հասցեներ,
- `.code`, `.data`, `.entry`,
- `.byte`, `.word`, `.string`, `.stringz`, `.zero`, `.align`,
- code/data label validation,
- EOF-ով ավարտվող տողի աջակցություն։

Չավելացնել macros, include, object/linker directives կամ function metadata։

Ավարտի չափանիշ՝ assembler-ի և Builder-ի համարժեք ծրագրերը տալիս են նույն `Program`-ը։

## Փուլ 5․ ֆունկցիաների ABI

- Պահպանել հիշողությունում տեսանելի call frame-ը։
- Արգումենտները push անել ձախից աջ։
- `CALL`-ով պահել վերադարձի `IP`-ն և հին `FP`-ը։
- Լոկալները հատկացնել `PUSH 0` հրահանգներով։
- Իրականացնել `RET n`, որը վերադարձնում է մեկ արժեք և հեռացնում `n` արգումենտ։
- Սահմանել `RET` որպես `RET 0`։
- Void ֆունկցիաներին պարտադրել վերադարձնել `0`։

Ավարտի չափանիշ՝ պարզ, nested, մի քանի արգումենտով և recursive կանչերի թեստերն անցնում են։

## Փուլ 6․ Machine runtime

Առաջարկվող API-ն՝

```go
machine, err := machine.New(machine.Config{
    MemorySize: 16 * 1024,
    Input:      input,
    Output:     output,
})

if err := machine.Load(program); err != nil {
    // invalid program
}
if err := machine.Run(); err != nil {
    // runtime error
}
```

Պահանջվող հնարավորությունները՝

- configurable, առավելագույնը 64 ԿԲ հիշողություն,
- `Load`, `Reset`, `Run`, `Step`,
- `IP`, `SP`, `FP` վիճակի read-only դիտարկում,
- ներարկվող input/output հոսքեր,
- հիշողության և ստեկի սահմանների ստուգում,
- անվավեր opcode/mode/branch target-ի ստուգում,
- զրոյի վրա բաժանման և I/O սխալների վերադարձ,
- պարզ `RuntimeError{IP, Message}`,
- սովորական ծրագրային սխալների դեպքում panic-ի բացակայություն։

Optional step limit-ը կարելի է ավելացնել որպես `RunOptions`, եթե անվերջ ցիկլով թեստերի համար անհրաժեշտ լինի։

## Փուլ 7․ executable format և CLI

`.svm` ֆայլը կունենա 12-բայթանոց header՝

```text
magic/version  4 բայթ
entry          2 բայթ
code_size      2 բայթ
data_size      2 բայթ
flags          2 բայթ
```

Ֆայլում header-ից հետո գրվում են code, zero padding և data բայթերը։ Symbol/debug table կամ checksum չկա։

CLI հրամանները՝

```sh
svm asm program.asm -o program.svm
svm run program.svm
svm exec program.asm
svm disasm program.svm
```

Օգտակար option-ներ՝ `--memory`, `--trace`, `--input`։ CLI-ն պետք է ունենա stdout/stderr-ի հստակ տարանջատում և կանխատեսելի exit code-եր։

## Փուլ 8․ թեստավորում և փաստաթղթեր

Թեստերը պետք է ներառեն՝

- յուրաքանչյուր instruction-ի unit test,
- encoder/decoder round-trip,
- assembler golden և malformed-input թեստեր,
- Builder-vs-assembler equivalence,
- memory/stack/runtime error թեստեր,
- ABI contract և recursion թեստեր,
- executable decode-ի կտրված/սխալ input-ներ,
- end-to-end օրինակներ։

Պարտադիր օրինակ ծրագրերը՝

- թվաբանական հաշվարկ,
- `if` և `while`,
- երկու թվերից մեծը,
- recursive factorial,
- global counter,
- array/string traversal,
- input/output։

README-ն պետք է պարունակի build/run արագ մեկնարկ, իսկ `docs/`-ը՝ հաստատված machine/ISA/ABI/assembler reference-ը։

## Առաջին թողարկման սահմանները

SVM v1.0-ում չկան՝

- JIT և floating-point,
- heap/GC,
- threads,
- memory permissions,
- syscall կամ plugin ABI,
- `.svo` object files և linker,
- `.global`/`.extern`,
- symbol/debug section-ներ,
- interactive debugger,
- macros և includes։

## Իրականացման հերթականությունը

```text
specification review
  → instruction encoding և Program
  → Builder
  → assembler
  → RET n ABI
  → Machine runtime
  → executable/CLI
  → end-to-end tests և փաստաթղթեր
```

Յուրաքանչյուր փուլ պետք է ավարտվի անցնող թեստերով և առանձին ստուգելի արդյունքով։
