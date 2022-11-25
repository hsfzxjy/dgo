part of dgo;

class DgoTypeLoadResult<Target> {
  final int nextIndex;
  final Target result;

  const DgoTypeLoadResult(this.nextIndex, this.result);
}

typedef DgoTypeLoader<Target> = DgoTypeLoadResult<Target> Function(
    List<dynamic> args, int startIndex);
