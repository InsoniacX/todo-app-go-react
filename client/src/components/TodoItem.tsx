import { Badge, Box, Flex, Spinner, Text } from "@chakra-ui/react";
import { FaCheckCircle } from "react-icons/fa";
import { MdDelete } from "react-icons/md";
import type { Todo } from "./TodoList";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { BASE_URL } from "../App";

const TodoItem = ({ todo }: { todo: Todo }) => {
  const queryClient = useQueryClient();

  // console.log(todo);

  const { mutate: updateTodo, isPending: isUpdating } = useMutation({
    mutationKey: ["updateTodo"],
    mutationFn: async () => {
      if (todo.completed) return alert("Todo is already completed!");
      try {
        const res = await fetch(BASE_URL + `/todos/${todo._id}`, {
          method: "PATCH",
        });

        if (!res.ok) {
          const text = await res.text();
          throw new Error(text || "Request failed");
        }

        return res.json().catch(() => null);
      } catch (error) {
        throw new Error(
          error instanceof Error ? error.message : "Unknown error"
        );
      }
    },

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["todos"] });
    },
    onError: (error) => {
      console.error(error);
    },
  });

  const { mutate: deleteTodo, isPending: isDeleting } = useMutation({
    mutationKey: ["deleteTodo"],
    mutationFn: async () => {
      try {
        const res = await fetch(BASE_URL + `/todos/${todo._id}`, {
          method: "DELETE",
        });
        const data = await res.json();
        if (!res.ok) {
          throw new Error(data.error || "Something went wrong");
        }
      } catch (error) {
        throw new Error(
          error instanceof Error ? error.message : "Unknown error"
        );
      }
    },

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["todos"] });
    },
  });

  return (
    <Flex gap={2} alignItems={"center"}>
      <Flex
        flex={1}
        alignItems={"center"}
        border={"1px"}
        borderColor={"gray.600"}
        p={2}
        borderRadius={"lg"}
        justifyContent={"space-between"}
      >
        <Text
          color={todo.completed ? "green.200" : "yellow.100"}
          textDecoration={todo.completed ? "line-through" : "none"}
        >
          {todo.body}
        </Text>
        {todo.completed && (
          <Badge ml="1" colorScheme="green">
            Done
          </Badge>
        )}
        {!todo.completed && (
          <Badge ml="1" colorScheme="yellow">
            In Progress
          </Badge>
        )}
      </Flex>
      <Flex gap={2} alignItems={"center"}>
        <Box color={"green.500"} cursor={"pointer"}>
          {!isUpdating && (
            <FaCheckCircle size={25} onClick={() => updateTodo()} />
          )}
          {isUpdating && <Spinner size={"sm"} />}
        </Box>
        <Box color={"red.500"} cursor={"pointer"}>
          {!isDeleting && <MdDelete size={25} onClick={() => deleteTodo()} />}
          {isDeleting && <Spinner size={"sm"} />}
        </Box>
      </Flex>
    </Flex>
  );
};
export default TodoItem;
