part of dgo;

abstract class DgoObject {
  const DgoObject();

  @protected
  int $dgoStore(List<dynamic> args, int index);

  int get $dgoGoSize;
}

late final DgoObject Function(int typeId, Iterator args) _buildObjectById;
late final T Function<T extends DgoObject>(Iterator args) _buildObject;

@experimental
void registerTypes(
  DgoObject Function(int, Iterator) buildObjectById,
  T Function<T extends DgoObject>(Iterator) buildObject,
) {
  _buildObjectById = buildObjectById;
  _buildObject = buildObject;
}
