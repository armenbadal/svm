# SVM v1 executable ձևաչափ

Կարգավիճակ՝ **քննարկման նախագիծ 0.2**

`.svm` ֆայլը բաղկացած է 12-բայթանոց header-ից, code բայթերից, alignment padding-ից և data բայթերից։ Բոլոր բազմաբայթ թվերը little-endian են։

```text
┌──────────────────────┐
│ Header՝ 12 բայթ      │
├──────────────────────┤
│ Code                 │
├──────────────────────┤
│ Zero padding         │
├──────────────────────┤
│ Data                 │
└──────────────────────┘
```

## Header

| Offset | Չափ | Դաշտ | Նշանակություն |
|---:|---:|---|---|
| `0` | 4 | `magic` | `53 56 4d 01`, այսինքն՝ `SVM` և version `1` |
| `4` | 2 | `entry` | code-ի ներսում առաջին հրահանգի հասցե |
| `6` | 2 | `code_size` | code-ի չափը բայթերով |
| `8` | 2 | `data_size` | data-ի չափը բայթերով |
| `10` | 2 | `flags` | v1.0-ում պարտադիր `0` |

Major format version-ը magic-ի չորրորդ բայթն է։ Header-ի ապագա անհամատեղելի փոփոխությունը պահանջում է նոր version։ v1.0 loader-ը մերժում է ոչ զրոյական `flags` դաշտը։

## Body

Header-ից հետո անմիջապես գրված են `code_size` code բայթերը։ Դրանց հետևում մինչև 4-բայթանոց սահմանը գրվում են զրո padding բայթեր, ապա՝ `data_size` data բայթերը։

```text
data_base = align_up(code_size, 4)
file_size = 12 + data_base + data_size
```

Code-ի VM հասցեները սկսվում են `0`-ից։ Data-ի առաջին բայթի VM հասցեն `data_base` է։ Header-ի 12 բայթը հասցեային տարածքի մաս չէ։

`code_size`-ը պետք է մեծ լինի `0`-ից։ `entry < code_size` և entry-ն պետք է ցույց տա հրահանգի առաջին բայթը։ `data_base + data_size` արժեքը չպետք է գերազանցի 65536-ը կամ runtime-ի տրամադրած հիշողության չափը։

## Ինչ չկա ֆայլում

SVM v1.0 executable-ը չի պարունակում՝

- section table,
- symbol table,
- debug metadata,
- relocation-ներ,
- checksum,
- timestamp կամ build ID։

Assembler-ը կամ Builder-ը բոլոր label-ներն ու հասցեները լուծում է մինչև executable-ը գրելն ու վերադարձնում պատրաստ code/data պատկեր։ Debugger-ը կարող է առանձին in-memory symbol map ստանալ assembler-ից, բայց այն `.svm` format-ի մաս չէ։

## Ստուգում

Decoder-ը հերթականությամբ՝

1. ստուգում է 12 header բայթերի առկայությունը,
2. ստուգում է magic/version-ը և `flags == 0` պայմանը,
3. overflow-ից պաշտպանված ձևով հաշվում է սպասվող file size-ը,
4. պահանջում է file size-ի ճշգրիտ համապատասխանություն՝ առանց ավելորդ պոչի,
5. ստուգում է code/data memory սահմանները,
6. decode է անում ամբողջ instruction stream-ը,
7. ստուգում է entry point-ը և static branch target-ները,
8. միայն հաջողությունից հետո վերադարձնում է `Program`։

Անվավեր ֆայլը վերադարձնում է սովորական error և երբեք մասնակի `Program` կամ Go panic չի տալիս։

## Deterministic encoding

Նույն `Program`-ի encoding-ը միշտ պետք է տա նույն բայթերը։ Padding-ը և `flags`-ը գրվում են զրոներով։ Executable-ում host path, ժամ կամ պատահական տվյալ չի գրվում։

## v0 համատեղելիություն

v0-ն header չունեցող հում instruction stream է։ v1 loader-ը այն ինքնաբերաբար չի ընդունում։ Անհրաժեշտության դեպքում CLI-ն կարող է հետագայում ստանալ բացահայտ `--format=v0` compatibility option, բայց դա v1 specification-ի պարտադիր մաս չէ։
