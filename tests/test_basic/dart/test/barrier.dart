part of 'test_basic_dart_test.dart';

class Barrier {
  static final stages = <List<Completer<void>>>[];
  static final setups = <dynamic Function()>[];

  static void blockHere({required dynamic Function() setup}) {
    stages.add([]);
    setups.add(setup);
  }

  static Future<void> waitStage(int stageIndex) async {
    if (stageIndex > 0) {
      await Future.wait(stages[stageIndex - 1].map((c) => c.future));
    }
    await setups[stageIndex]();
  }

  static void test(dynamic description, dynamic Function() body) {
    final stageIndex = stages.length - 1;
    final c = Completer<void>();
    stages.last.add(c);
    t.test(description, () async {
      await waitStage(stageIndex);
      try {
        await body();
      } finally {
        c.complete();
      }
    });
  }
}

final test = Barrier.test;
