part of dgo;

abstract class Pinnable extends DgoObject {
  const Pinnable();
}

abstract class PinToken<T extends Pinnable> extends DgoObject {
  T dispose();
  @override
  final $dgoGoSize = 3;
}
