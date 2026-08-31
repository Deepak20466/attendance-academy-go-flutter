import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../data/student_repository.dart';

class StudentDetailScreen extends ConsumerWidget {
  const StudentDetailScreen({super.key, required this.studentId});
  final String studentId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final student = ref.watch(studentDetailProvider(studentId));
    final history = ref.watch(_historyProvider(studentId));

    return Scaffold(
      appBar: AppBar(title: const Text('Student')),
      body: AsyncValueWidget(
        value: student,
        data: (s) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Center(
              child: Column(
                children: [
                  CircleAvatar(radius: 32, child: Text(s.name.isNotEmpty ? s.name[0].toUpperCase() : '?')),
                  const SizedBox(height: 12),
                  Text(s.name, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w700)),
                  if (!s.isActive)
                    const Padding(
                      padding: EdgeInsets.only(top: 4),
                      child: Chip(label: Text('Inactive')),
                    ),
                ],
              ),
            ),
            const SizedBox(height: 20),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _row(Icons.phone_outlined, 'Phone', s.phone),
                    _row(Icons.family_restroom_outlined, 'Guardian', s.guardianName),
                    _row(Icons.call_outlined, 'Guardian phone', s.guardianPhone),
                    _row(Icons.email_outlined, 'Email', s.email),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 20),
            Text('Attendance history', style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700)),
            const SizedBox(height: 12),
            AsyncValueWidget(
              value: history,
              data: (result) {
                if (result.data.isEmpty) {
                  return const Padding(
                    padding: EdgeInsets.symmetric(vertical: 16),
                    child: Text('No attendance records yet.', style: TextStyle(color: Colors.grey)),
                  );
                }
                return Column(
                  children: result.data.map((entry) {
                    final color = switch (entry.status) {
                      'present' => Colors.green,
                      'late' => Colors.orange,
                      'excused' => Colors.grey,
                      _ => Colors.red,
                    };
                    return ListTile(
                      contentPadding: EdgeInsets.zero,
                      leading: CircleAvatar(backgroundColor: color.withValues(alpha: 0.15), child: Icon(Icons.circle, color: color, size: 12)),
                      title: Text(entry.status[0].toUpperCase() + entry.status.substring(1)),
                      subtitle: Text(DateFormat.yMMMd().add_jm().format(entry.markedAt)),
                    );
                  }).toList(),
                );
              },
            ),
          ],
        ),
      ),
    );
  }

  Widget _row(IconData icon, String label, String value) {
    if (value.isEmpty) return const SizedBox.shrink();
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        children: [
          Icon(icon, size: 18, color: Colors.grey),
          const SizedBox(width: 10),
          Text('$label: ', style: const TextStyle(color: Colors.grey)),
          Expanded(child: Text(value)),
        ],
      ),
    );
  }
}

final _historyProvider = FutureProvider.autoDispose.family((ref, String studentId) {
  return ref.watch(studentRepositoryProvider).history(studentId);
});
