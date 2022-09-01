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

  int _applyOn(int flag) {
    flag = flag & ~(_cfFast * 7); // 0b111. clear flag
    if (this == none) return flag;
    flag |= _cfFast;
    flag += (value - 1) << _cfBitsFastStart;
    return flag;
  }
}

const _cfBitsStart = 32;
const _cfBitsFastStart = _cfBitsStart + 4;

const _cfPop = 1 << (_cfBitsStart + 0);

const _cfWithCode = 1 << (_cfBitsStart + 1);
const _cfPackArray = 1 << (_cfBitsStart + 2);

const _cfFast = 1 << (_cfBitsStart + 3);
// ignore:unused_element
const _cfFastVoid = _cfFast + (0 << (_cfBitsStart + 4));
// ignore:unused_element
const _cfFastNil = _cfFast + (1 << (_cfBitsStart + 4));
// ignore:unused_element
const _cfFastYes = _cfFast + (2 << (_cfBitsStart + 4));
// ignore:unused_element
const _cfFastNo = _cfFast + (3 << (_cfBitsStart + 4));

// ignore:unused_element
const _cfFutReject = 0 << (_cfBitsStart + 6);
// ignore:unused_element
const _cfFutResolve = 1 << (_cfBitsStart + 6);

@immutable
class CallbackFlag {
  final int _internal;

  bool get hasPop => _internal & _cfPop != 0;
  bool get hasWithCode => _internal & _cfWithCode != 0;
  bool get hasPackArray => _internal & _cfPackArray != 0;
  bool get hasFast => fastKind != FastKind.none;
  FastKind get fastKind => FastKind._fromFlagInt(_internal);

  const CallbackFlag._(this._internal);
  const CallbackFlag() : this._(0);

  CallbackFlag pop() => CallbackFlag._(_internal | _cfPop);
  CallbackFlag withCode() => CallbackFlag._(_internal | _cfWithCode);
  CallbackFlag packArray() => CallbackFlag._(_internal | _cfPackArray);
  CallbackFlag fast(FastKind kind) => CallbackFlag._(kind._applyOn(_internal));

  bool isBitSet(int bitFlag) => bitFlag & _internal != 0;
}

//ignore: constant_identifier_names
const CF = CallbackFlag();
typedef CFFK = FastKind;
