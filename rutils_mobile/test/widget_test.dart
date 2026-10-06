import 'package:flutter_test/flutter_test.dart';
import 'package:rutils_mobile/main.dart';

void main() {
  testWidgets('RutilsApp boots smoke test', (WidgetTester tester) async {
    await tester.pumpWidget(const RutilsApp());
    await tester.pumpAndSettle(const Duration(seconds: 1));
    expect(find.byType(RutilsApp), findsOneWidget);
  });
}
