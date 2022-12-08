part of 'ir.dart';

class Param {
  final String name;
  final IR term;

  Param.fromMap(Map m)
      : term = _buildIR(m['Term']),
        name = m['Name'];
}

extension _NullableIRExt on IR? {
  String get dartType => this == null ? 'void' : this!.dartType;
}

extension _IRExt on IR {
  void writeSnippetLoad(GeneratorContext ctx) {
    if (this is Namable && (this as Namable).isNamed) {
      final myUri = (this as Namable).myUri!;
      ctx
        ..sln('{')
        ..sln('$vHolder = ')
        ..sln(ctx.importer.qualifyUri(myUri))
        ..sln('.\$dgoLoad($vArgs);')
        ..sln('}');
    } else {
      writeSnippet$dgoLoad();
    }
  }
}

extension _SumExt<T extends num> on Iterable<T> {
  T sum() => isEmpty ? 0 as T : reduce((a, b) => (a + b) as T);
}

class Method {
  final IR self;
  final String funcName;
  final List<Param> params;
  final int funcId;
  final IR? returnType;
  final bool returnError;

  Method.fromMap(this.self, Map m)
      : funcName = m['Name'],
        funcId = m['FuncId'],
        params =
            (m['Params'] as List).cast<JsonMap>().map(Param.fromMap).toList(),
        returnType = _buildIRNull(m['Return']),
        returnError = m['ReturnError'];

  bool get isGoNotDynamic =>
      self.isGoNotDynamic && params.every((p) => p.term.isGoNotDynamic);

  void _buildSize() {
    if (isGoNotDynamic) {
      final size = params.map((p) => p.term.goSize).sum() + self.goSize;
      ctx.sln('final $vSize = $size;');
      return;
    }

    ctx
      ..sln()
      ..sln('\$core.int $vSize = \$dgoGoSize;')
      ..for_(
        params,
        (p) => ctx
          ..if_(
            p.term.isGoNotDynamic,
            () => ctx.sln('$vSize += ${p.term.goSize};'),
            else_: () => ctx
              ..sln('{ final $vHolder = ${p.name};')
              ..then(p.term.writeSnippet$dgoGoSize)
              ..sln('}'),
          ),
      )
      ..sln();
  }

  String signature(String funcName) {
    var paramSig =
        params.map((p) => '${p.term.outerDartType} ${p.name}').followedBy([
      '{\$core.Duration? \$timeout, \$dgo.DgoPort? \$port}',
    ]).joinComma;
    return '\$async.Future<${returnType.dartType}> $funcName($paramSig)';
  }

  void writeSnippet() {
    ctx
      ..sln('${signature(funcName)} async {')
      ..sln('\$port ??= \$dgo.dgo.defaultPort;')
      ..then(_buildSize)
      ..sln('final $vArgs = \$core.List<\$core.dynamic>'
          '.filled($vSize, null, growable: false);')
      ..sln('var $vIndex = 0;')
      ..sln('$vIndex = \$dgoStore($vArgs, $vIndex);')
      ..for_(
          params,
          (param) => ctx
            ..sln('{')
            ..sln('final $vHolder = ${param.name};')
            ..then(param.term.writeSnippet$dgoStore)
            ..sln('}'))
      ..sln(
        'final \$future = \$dgo.GoMethod($funcId, \$port)'
        '.${returnType == null ? "call" : "callWithResult"}'
        '($vArgs, timeout: \$timeout, hasError: $returnError);',
      )
      ..if_(
        returnType == null,
        () => ctx.sln('return \$future;'),
        else_: () => ctx
          ..sln('{')
          ..sln('final $vArgs = (await \$future).iterator;')
          ..sln('$vArgs.moveNext();')
          ..sln('${returnType!.dartType} $vHolder;')
          ..pipe(returnType!.writeSnippetLoad(ctx))
          ..sln('return $vHolder;')
          ..sln('}'),
      )
      ..sln('}');
  }
}
