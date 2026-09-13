# SVM v1 ֆունկցիաների ABI

Կարգավիճակ՝ **քննարկման նախագիծ 0.2**

SVM v1-ի ֆունկցիաները օգտագործում են նույն ընդհանուր ստեկը։ Function table, descriptor կամ առանձին local storage չկա։ Call frame-ը ամբողջությամբ տեսանելի է հիշողությունում։

## Կանոններ

- Յուրաքանչյուր արգումենտ, լոկալ և վերադարձվող արժեք զբաղեցնում է 4 բայթ։
- Caller-ը արգումենտները push է անում ձախից աջ։
- `CALL`-ը պահպանում է վերադարձի `IP`-ն ու caller-ի `FP`-ը։
- Callee-ն լոկալները հատկացնում է `PUSH 0` հրահանգներով։
- Callee-ն ստեկի գագաթին թողնում է ճիշտ մեկ վերադարձվող արժեք։
- `RET n`-ը վերադարձնում է արժեքը և հեռացնում է caller-ի `n` արգումենտները։
- Void ֆունկցիան նույնպես վերադարձնում է մեկ `0` արժեք։

Multiple return value և variadic function v1.0-ում չկան։ Մեծ կառուցվածքները փոխանցվում են հասցեով։

## `CALL`-ի վարքը

Թող `P`-ն լինի `SP`-ն վերջին արգումենտից հետո։ `CALL target`-ը՝

1. ստուգում է target-ը և երկու ազատ stack slot-ի առկայությունը,
2. `[P]` հասցեում գրում է հաջորդ instruction-ի հասցեն,
3. `[P + 4]` հասցեում գրում է հին `FP`-ն,
4. սահմանում է `FP = P + 8`,
5. սահմանում է `SP = FP`,
6. սահմանում է `IP = target`։

Frame-ի տեսքը՝

```text
փոքր հասցեներ

... caller-ի արժեքներ ...
[FP - 8 - 4*N] arg0
...
[FP - 16]       argN-2
[FP - 12]       argN-1
[FP - 8]        վերադարձի IP
[FP - 4]        հին FP
[FP + 0]        local0 կամ առաջին operand
[FP + 4]        local1
...

մեծ հասցեներ
```

`N` արգումենտ ունեցող ֆունկցիայի `argI` հասցեն՝

```text
FP - 8 - 4 * (N - I)
```

Վերջին արգումենտը միշտ `[FP - 12]` հասցեում է։

## Լոկալներ

Լոկալը հատկացնելն ու զրոյացնելը նույն գործողությունն է՝

```asm
function:
    PUSH 0      ; local0՝ [FP + 0]
    PUSH 0      ; local1՝ [FP + 4]
```

Լոկալը կարդալու և գրելու համար օգտագործվում են սովորական հարաբերական հրահանգները․

```asm
PUSH [FP + 0]

PUSH 42
POP [FP + 4]
```

Compiler-ը կարող է մեկ `PUSH 0` գեներացնել յուրաքանչյուր local slot-ի համար։ Assembler-ը հետագայում կարող է ունենալ `.locals n` հարմարության directive, բայց դա VM instruction չէ։

Callee-ն չպետք է գրի `[FP - 8]` կամ `[FP - 4]` ծառայողական slot-ներում։

## `RET n`-ի վարքը

`n`-ը caller-ի մաքրվող արգումենտների քանակն է, ոչ բայթերի թիվը։ `RET` առանց operand-ի համարժեք է `RET 0`-ին։

Թող՝

```text
result        = memory32[SP - 4]
return_ip     = memory32[FP - 8]
caller_fp     = memory32[FP - 4]
arguments_end = FP - 8
caller_sp     = arguments_end - 4*n
```

Բոլոր սահմանները և frame արժեքները ստուգելուց հետո VM-ը՝

```text
IP = return_ip
FP = caller_fp
SP = caller_sp
PUSH(result)
```

Այսպիսով callee-ի լոկալները, ժամանակավոր արժեքները և caller-ի փոխանցած `n` արգումենտները հեռացվում են, իսկ caller-ը ստանում է միայն արդյունքը։

```text
... arg0 arg1    CALL-ից առաջ
... result       RET 2-ից հետո
```

Եթե `n`-ը կփորձի մաքրել caller-ի հասանելի ստեկից ավելի շատ արժեք, առաջանում է runtime error, և frame-ը մասնակի չի վերականգնվում։ Direct call-ի արգումենտների ճիշտ քանակը հիմնականում compiler/assembler ծրագրի պատասխանատվությունն է։

## Entry point

Ծրագրի entry point-ը ֆունկցիայի սովորական կանչ չէ։ Loader-ը սկսում է՝

```text
IP = entry
SP = stack_base
FP = 0
```

Entry point-ը չի կարող ավարտվել `RET`-ով։ Այն պետք է ավարտվի `HALT`-ով։ Սովորական bootstrap-ը՝

```asm
.entry start
.code

start:
    CALL main
    PRINT
    HALT

main:
    PUSH 0
    RET
```

## Օրինակ՝ երկու թվերից մեծը

```asm
.entry start
.code

start:
    PUSH 10
    PUSH 20
    CALL max
    PRINT
    HALT

max:
    PUSH [FP - 16]    ; առաջին արգումենտը
    PUSH [FP - 12]    ; երկրորդ արգումենտը
    GT
    JZ second
    PUSH [FP - 16]
    RET 2

second:
    PUSH [FP - 12]
    RET 2
```

`CALL max`-ից հետո արգումենտներն այլևս ստեկում չեն, իսկ `PRINT`-ը ստանում է միայն `20` արդյունքը։

## Օրինակ՝ recursion

```asm
; factorial(n)
factorial:
    PUSH [FP - 12]
    PUSH 1
    LE
    JZ recurse
    PUSH 1
    RET 1

recurse:
    PUSH [FP - 12]
    PUSH 1
    SUB
    CALL factorial
    PUSH [FP - 12]
    MUL
    RET 1
```

Recursive `RET 1`-ը մաքրում է տվյալ կանչի մեկ արգումենտը։ Արտաքին frame-ի `FP`-ն և արժեքները պահպանվում են։
