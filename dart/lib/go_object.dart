part of dgo;

abstract class DgoObject {
  const DgoObject();

  @visibleForOverriding
  int $dgoStore(List<dynamic> args, int index);

  int get $dgoGoSize;
}

late final DgoObject Function(int typeId, DgoPort port, Iterator args)
    _buildObjectById;
late final T Function<T extends DgoObject>(DgoPort port, Iterator args)
    _buildObject;

@experimental
void registerTypes(
  DgoObject Function(int, DgoPort, Iterator) buildObjectById,
  T Function<T extends DgoObject>(DgoPort, Iterator) buildObject,
) {
  _buildObjectById = buildObjectById;
  _buildObject = buildObject;
}
