import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../../auth/data/auth_controller.dart';
import '../../classes/data/class_repository.dart';
import '../../coaches/data/coach_repository.dart';

class CoachDashboardScreen extends ConsumerWidget {
  const CoachDashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(authControllerProvider).valueOrNull;
    final profile = ref.watch(myCoachProfileProvider);
    final todayClasses = ref.watch(todayClassesProvider);

    return Scaffold(
      appBar: AppBar(title: Text('Hi, ${session?.name.isNotEmpty == true ? session!.name : 'Coach'}')),
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(myCoachProfileProvider);
          ref.invalidate(todayClassesProvider);
        },
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            AsyncValueWidget(
              value: profile,
              data: (p) => Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Row(
                    children: [
                      CircleAvatar(radius: 24, child: Text(p.name.isNotEmpty ? p.name[0].toUpperCase() : '?')),
                      const SizedBox(width: 14),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(p.name, style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 16)),
                            Text('Employee code: ${p.employeeCode}', style: const TextStyle(color: Colors.grey)),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
            const SizedBox(height: 20),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text("Today's classes", style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700)),
              ],
            ),
            const SizedBox(height: 12),
            AsyncValueWidget(
              value: todayClasses,
              data: (classes) {
                if (classes.isEmpty) {
                  return const Padding(
                    padding: EdgeInsets.symmetric(vertical: 24),
                    child: Center(child: Text('No classes scheduled today.', style: TextStyle(color: Colors.grey))),
                  );
                }
                return Column(
                  children: classes.map((c) {
                    return Card(
                      margin: const EdgeInsets.only(bottom: 10),
                      child: ListTile(
                        leading: Icon(
                          c.attendanceMarked ? Icons.check_circle : Icons.schedule,
                          color: c.attendanceMarked ? Colors.green : Colors.orange,
                        ),
                        title: Text('${c.startTime} - ${c.endTime}'),
                        subtitle: Text(c.attendanceMarked ? 'Attendance submitted' : 'Attendance pending'),
                        trailing: Wrap(
                          spacing: 4,
                          children: [
                            IconButton(
                              icon: const Icon(Icons.location_on_outlined),
                              tooltip: 'Check in / out',
                              onPressed: () async {
                                await context.push('/coach-checkin/${c.id}');
                                ref.invalidate(todayClassesProvider);
                              },
                            ),
                            IconButton(
                              icon: const Icon(Icons.checklist_outlined),
                              tooltip: 'Mark attendance',
                              onPressed: () async {
                                await context.push('/classes/${c.id}/attendance');
                                ref.invalidate(todayClassesProvider);
                              },
                            ),
                          ],
                        ),
                      ),
                    );
                  }).toList(),
                );
              },
            ),
            const SizedBox(height: 80),
          ],
        ),
      ),
    );
  }
}
