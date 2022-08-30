part of dgo;

enum FastKind {
  none(0),
  void_(1),
  nil(2),
  yes(3),
  no(4);

  const FastKind(this.value);

  final int value;

  factory FastKind._fromFlagInt(int x) {
    if (x & (1 << (_cfBitsStart + 3)) == 0) return none;
    x >>>= _cfBitsStart + 4;
    x &= 3; // 0b11
    switch (x) {
      case 0:
        return void_;
      case 1:
        return nil;
      case 2:
        return yes;
      case 3:
        return no;
      default:
        throw 'unreachable';
    }
  }
}

const _cfBitsStart = 32;

@immutable
class CallbackFlag {
  final bool hasPop;
  final bool hasWithCode;
  final bool hasPackArray;
  final FastKind fastKind;

  bool get hasFast => fastKind != FastKind.none;

  const CallbackFlag._(
      this.hasPop, this.hasWithCode, this.hasPackArray, this.fastKind);
  const CallbackFlag() : this._(false, false, false, FastKind.none);

  CallbackFlag pop() =>
      CallbackFlag._(true, hasWithCode, hasPackArray, fastKind);
  CallbackFlag withCode() =>
      CallbackFlag._(hasPop, true, hasPackArray, fastKind);
  CallbackFlag packArray() =>
      CallbackFlag._(hasPop, hasWithCode, true, fastKind);
  CallbackFlag fast(FastKind kind) =>
      CallbackFlag._(hasPop, hasWithCode, hasPackArray, kind);

  int _asInt() {
    int x = 0;
    if (hasPop) x |= 1 << (_cfBitsStart + 0);
    if (hasWithCode) x |= 1 << (_cfBitsStart + 1);
    if (hasPackArray) x |= 1 << (_cfBitsStart + 2);
    switch (fastKind) {
      case FastKind.none:
        break;
      default:
        x |= 1 << (_cfBitsStart + 3);
        x += (fastKind.value - 1) << (_cfBitsStart + 4);
    }
    return x;
  }

  factory CallbackFlag._fromInt(int x) {
    final hasPop = x & (1 << (_cfBitsStart + 0)) != 0;
    final hasWithCode = x & (1 << (_cfBitsStart + 1)) != 0;
    final hasPackArray = x & (1 << (_cfBitsStart + 2)) != 0;
    final fastKind = FastKind._fromFlagInt(x);
    return CallbackFlag._(hasPop, hasWithCode, hasPackArray, fastKind);
  }
}

//ignore: constant_identifier_names
const CF = CallbackFlag();
typedef CFFK = FastKind;
