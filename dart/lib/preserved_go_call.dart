part of dgo;

@experimental
enum PreservedGoCall implements _Serializable {
  tokenDispose(1),
  chanListen(2),
  chanCancelListen(3);

  const PreservedGoCall(this._payload);
  @override
  final int _payload;

  @override
  _SpecialIntKind get _kind => _SpecialIntKind.prevseredGoCall;

  void call(DgoPort port, Iterable<dynamic> args) =>
      port._postList(_OneShot<dynamic>(this).followedBy(args));
}
